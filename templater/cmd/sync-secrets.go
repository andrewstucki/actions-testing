// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package cmd

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	gogithub "github.com/google/go-github/v69/github"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/nacl/box"
	"gopkg.in/yaml.v3"

	"github.com/andrewstucki/actions-testing/templater/config"
	"github.com/andrewstucki/actions-testing/templater/github"
	"github.com/andrewstucki/actions-testing/templater/prompt"
)

// syncSecretsCmd represents the sync-secrets command
var syncSecretsCmd = &cobra.Command{
	Use:   "sync-secrets",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("error reading configuration file: %v\n", err)
			os.Exit(1)
		}

		var cfg config.ConfigFile
		err = yaml.Unmarshal(data, &cfg)
		if err != nil {
			fmt.Printf("error unmarshaling configuration file: %v\n", err)
			os.Exit(1)
		}

		organization, repo := cfg.GithubInfo.Organization, cfg.GithubInfo.Repository
		client, err := github.Client()
		if err != nil {
			fmt.Printf("error getting Github client: %v\n", err)
			os.Exit(1)
		}

		token, webhook, confirmed, err := prompt.RunSecretSync(cfg)
		if err != nil {
			fmt.Printf("error getting secrets: %v\n", err)
			os.Exit(1)
		}
		if !confirmed {
			fmt.Print("sync canceled\n")
			os.Exit(1)
		}

		pubKey, _, err := client.Actions.GetRepoPublicKey(cmd.Context(), organization, repo)
		if err != nil {
			fmt.Printf("error fetching encryption key: %v\n", err)
			os.Exit(1)
		}
		decodedPubKey, err := base64.StdEncoding.DecodeString(pubKey.GetKey())
		if err != nil {
			fmt.Printf("error decoding encryption key: %v\n", err)
			os.Exit(1)
		}
		var peersPubKey [32]byte
		copy(peersPubKey[:], decodedPubKey[0:32])

		encryptAndSet := func(name, value string) error {
			value = strings.TrimSpace(value)
			if value == "" {
				return nil
			}

			var rand io.Reader
			encryptedBody, err := box.SealAnonymous(nil, []byte(value)[:], &peersPubKey, rand)
			if err != nil {
				return err
			}

			encoded := base64.StdEncoding.EncodeToString(encryptedBody)
			_, err = client.Actions.CreateOrUpdateRepoSecret(cmd.Context(), organization, repo, &gogithub.EncryptedSecret{
				Name:           name,
				EncryptedValue: encoded,
				KeyID:          pubKey.GetKeyID(),
			})
			return err
		}

		if err := encryptAndSet(cfg.Backports.Bot.TokenVariable, token); err != nil {
			fmt.Printf("error setting %q: %v\n", cfg.Backports.Bot.TokenVariable, err)
			os.Exit(1)
		}

		if err := encryptAndSet("SLACK_WEBHOOK_URL", webhook); err != nil {
			fmt.Printf("error setting \"SLACK_WEBHOOK_URL\": %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(syncSecretsCmd)
}
