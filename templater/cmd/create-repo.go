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
)

// createRepoCmd represents the create-repo command
var createRepoCmd = &cobra.Command{
	Use:   "create-repo",
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

		client, err := github.GetClient()
		if err != nil {
			fmt.Printf("error getting Github client: %v\n", err)
			os.Exit(1)
		}

		_, err = client.InitializeRepository(cmd.Context(), cfg.GithubInfo.Organization, cfg.GithubInfo.Repository)
		if err != nil {
			fmt.Printf("error initializing Github repo: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(createRepoCmd)
}
