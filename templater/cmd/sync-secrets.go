// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
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

		secrets, confirmed, err := prompt.RunSecretSync(cfg)
		if err != nil {
			fmt.Printf("error getting secrets: %v\n", err)
			os.Exit(1)
		}
		if !confirmed {
			fmt.Print("sync canceled\n")
			os.Exit(1)
		}

		client, err := github.GetRepoClient(cmd.Context(), cfg.GithubInfo.Organization, cfg.GithubInfo.Repository)
		if err != nil {
			fmt.Printf("error getting Github client: %v\n", err)
			os.Exit(1)
		}

		for _, secret := range secrets {
			if err := client.SetEncryptedSecret(cmd.Context(), secret.Name, secret.Value); err != nil {
				fmt.Printf("error setting %q: %v\n", secret.Name, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(syncSecretsCmd)
}
