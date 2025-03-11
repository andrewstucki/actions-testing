package github

import (
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/zalando/go-keyring"
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

func Client() (*github.Client, error) {
	_, token, err := getUserTokenForGithub()
	if err != nil {
		return nil, err
	}

	client := github.NewClient(nil).WithAuthToken(token)
	return client, nil
}
