package github

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/nacl/box"
	"gopkg.in/yaml.v3"
)

var ErrNotFound = errors.New("secret not found in keyring")

type TimeoutError struct {
	message string
}

func (e *TimeoutError) Error() string {
	return e.message
}

// from https://github.com/cli/cli/blob/af4acb380136fd106b38cc0ab404ff975bca9795/internal/keyring/keyring.go#L37C1-L59C2
func getToken(service, user string) (string, error) {
	ch := make(chan struct {
		val string
		err error
	}, 1)
	go func() {
		defer close(ch)
		val, err := keyring.Get(service, user)
		ch <- struct {
			val string
			err error
		}{val, err}
	}()
	select {
	case res := <-ch:
		if errors.Is(res.err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return res.val, res.err
	case <-time.After(3 * time.Second):
		return "", &TimeoutError{"timeout while trying to get secret from keyring"}
	}
}

type host struct {
	User string `yaml:"user"`
}

func getGithubUser() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	configFile := path.Join(home, ".config", "gh", "hosts.yml")
	data, err := os.ReadFile(configFile)
	if err != nil {
		return "", fmt.Errorf("reading config file: %w", err)
	}

	hosts := map[string]host{}
	if err := yaml.Unmarshal(data, hosts); err != nil {
		return "", fmt.Errorf("unmarshaling config file: %w", err)
	}

	github, ok := hosts["github.com"]
	if !ok || github.User == "" {
		return "", errors.New("unable to find active github user")
	}

	return github.User, nil
}

func getUserTokenForGithub() (string, string, error) {
	user, err := getGithubUser()
	if err != nil {
		return "", "", err
	}

	token, err := getToken("gh:github.com", user)
	if err != nil {
		return "", "", fmt.Errorf("fetching token: %w", err)
	}

	return user, token, nil
}

type Client struct {
	*github.Client

	user string

	organization  string
	repo          string
	peerPublicKey [32]byte
	key           *github.PublicKey
}

func GetClient() (*Client, error) {
	user, token, err := getUserTokenForGithub()
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(token)
	return &Client{Client: client, user: user}, nil
}

func GetRepoClient(ctx context.Context, organization, repo string) (*Client, error) {
	client, err := GetClient()
	if err != nil {
		return nil, err
	}

	return client.SetRepository(ctx, organization, repo)
}

func (c *Client) InitializeRepository(ctx context.Context, organization, repo string) (string, error) {
	userOrganization := organization
	if c.user == organization {
		userOrganization = ""
	}

	createdRepo, _, err := c.Client.Repositories.Create(ctx, userOrganization, &github.Repository{
		Name:                github.Ptr(repo),
		AllowAutoMerge:      github.Ptr(true),
		DeleteBranchOnMerge: github.Ptr(true),
		HasWiki:             github.Ptr(false),
		HasProjects:         github.Ptr(false),
	})
	if err != nil {
		return "", err
	}

	_, _, err = c.Client.Repositories.CreateRuleset(ctx, organization, repo, github.RepositoryRuleset{
		Name:        "Require PR",
		Enforcement: github.RulesetEnforcementActive,
		Conditions: &github.RepositoryRulesetConditions{
			RefName: &github.RepositoryRulesetRefConditionParameters{
				Include: []string{"~DEFAULT_BRANCH", "refs/heads/v**"},
				Exclude: []string{},
			},
		},
		Rules: &github.RepositoryRulesetRules{
			Deletion:       &github.EmptyRuleParameters{},
			NonFastForward: &github.EmptyRuleParameters{},
			PullRequest: &github.PullRequestRuleParameters{
				AllowedMergeMethods: []github.MergeMethod{
					github.MergeMethodMerge, github.MergeMethodRebase, github.MergeMethodSquash,
				},
				DismissStaleReviewsOnPush:      false,
				RequireCodeOwnerReview:         false,
				RequireLastPushApproval:        false,
				RequiredApprovingReviewCount:   1,
				RequiredReviewThreadResolution: false,
			},
		},
	})
	if err != nil {
		return "", err
	}

	// c.Client.Repositories.UpdateRequiredStatusChecks()

	_, _, err = c.Client.Repositories.EditDefaultWorkflowPermissions(ctx, organization, repo, github.DefaultWorkflowPermissionRepository{
		DefaultWorkflowPermissions:   github.Ptr("write"),
		CanApprovePullRequestReviews: github.Ptr(true),
	})
	if err != nil {
		return "", err
	}

	return createdRepo.GetSSHURL(), nil
}

func (c *Client) SetRepository(ctx context.Context, organization, repo string) (*Client, error) {
	var err error

	c.organization, c.repo = organization, repo
	c.key, c.peerPublicKey, err = c.getPublicKey(ctx)
	return c, err
}

func (c *Client) SetEncryptedSecret(ctx context.Context, name, value string) error {
	if c.organization == "" || c.repo == "" {
		return errors.New("must set repository before setting encrypted secret")
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	var rand io.Reader
	encryptedBody, err := box.SealAnonymous(nil, []byte(value)[:], &c.peerPublicKey, rand)
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(encryptedBody)
	_, err = c.Client.Actions.CreateOrUpdateRepoSecret(ctx, c.organization, c.repo, &github.EncryptedSecret{
		Name:           name,
		EncryptedValue: encoded,
		KeyID:          c.key.GetKeyID(),
	})
	return err
}

func (c *Client) getPublicKey(ctx context.Context) (*github.PublicKey, [32]byte, error) {
	var peerPubKey [32]byte
	pubKey, _, err := c.Client.Actions.GetRepoPublicKey(ctx, c.organization, c.repo)
	if err != nil {
		return nil, peerPubKey, err
	}
	decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey.GetKey())
	if err != nil {
		return nil, peerPubKey, err
	}
	copy(peerPubKey[:], decodedPubKey[0:32])

	return pubKey, peerPubKey, nil
}
