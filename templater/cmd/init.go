// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrewstucki/actions-testing/templater/github"
	"github.com/andrewstucki/actions-testing/templater/prompt"
	"github.com/andrewstucki/actions-testing/templater/templates"
)

var (
	skipTidy       bool
	initializeRepo bool
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "A brief description of your command",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := prompt.Run()
		if err != nil {
			fmt.Printf("error initializing project: %v\n", err)
			os.Exit(1)
		}

		data, err := yaml.Marshal(cfg)
		if err != nil {
			fmt.Printf("error marshaling config: %v\n", err)
			os.Exit(1)
		}

		if _, err := os.Stat(cfg.GithubInfo.Repository); os.IsExist(err) {
			fmt.Printf("cannot initialize project in %q, folder already exists\n", cfg.GithubInfo.Repository)
			os.Exit(1)
		}

		info := templates.TemplateInfo{
			Copyright:            cfg.License.Copyright,
			License:              cfg.License.License,
			Organization:         cfg.GithubInfo.Organization,
			Repository:           cfg.GithubInfo.Repository,
			BackportBranches:     cfg.Backports.Branches,
			Versions:             cfg.Backports.Versions,
			BackportBot:          cfg.Backports.Bot.Name,
			BackportBotTokenVar:  cfg.Backports.Bot.TokenVariable,
			Label:                cfg.Backports.Label,
			LabelMapper:          cfg.Backports.Mappings,
			LicenseManagement:    true,
			Backports:            true,
			AutoApproveBackports: true,
		}

		for _, project := range cfg.Projects {
			info.Projects = append(info.Projects, templates.ProjectInfo{
				Name:      project.Name,
				Changelog: project.Changelog,
			})
		}

		if err := templates.RenderTo(cfg.GithubInfo.Repository, info); err != nil {
			fmt.Printf("error rendering templates: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(path.Join(cfg.GithubInfo.Repository, ".template.yaml"), data, 0644); err != nil {
			fmt.Printf("error writing config file: %v\n", err)
			os.Exit(1)
		}

		if err := os.Chdir(cfg.GithubInfo.Repository); err != nil {
			fmt.Printf("error changing directory: %v\n", err)
			os.Exit(1)
		}

		if _, err := exec.Command("git", "init").CombinedOutput(); err != nil {
			fmt.Printf("error running git: %v\n", err)
			os.Exit(1)
		}

		if !skipTidy {
			if _, err := exec.Command("direnv", "allow").CombinedOutput(); err != nil {
				fmt.Printf("error running direnv: %v\n", err)
				os.Exit(1)
			}

			if _, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
				fmt.Printf("error running go mod tidy: %v\n", err)
				os.Exit(1)
			}

			if _, err := exec.Command("git", "add", ".").CombinedOutput(); err != nil {
				fmt.Printf("error running git: %v\n", err)
				os.Exit(1)
			}

			if _, err := exec.Command("nix", "develop", "-c", "licenseupdater").CombinedOutput(); err != nil {
				fmt.Printf("error running licenseupdater: %v\n", err)
				os.Exit(1)
			}

			if _, err := exec.Command("nix", "develop", "-c", "changie", "merge").CombinedOutput(); err != nil {
				fmt.Printf("error running changie: %v\n", err)
				os.Exit(1)
			}
		}

		if _, err := exec.Command("git", "add", ".").CombinedOutput(); err != nil {
			fmt.Printf("error running git: %v\n", err)
			os.Exit(1)
		}

		if _, err := exec.Command("git", "commit", "-m", "initial commit").CombinedOutput(); err != nil {
			fmt.Printf("error running git: %v\n", err)
			os.Exit(1)
		}

		if initializeRepo {
			client, err := github.GetClient()
			if err != nil {
				fmt.Printf("error getting Github client: %v\n", err)
				os.Exit(1)
			}

			url, err := client.InitializeRepository(cmd.Context(), cfg.GithubInfo.Organization, cfg.GithubInfo.Repository)
			if err != nil {
				fmt.Printf("error initializing Github repo: %v\n", err)
				os.Exit(1)
			}

			secrets, confirmed, err := prompt.RunSecretSync(*cfg)
			if err != nil {
				fmt.Printf("error getting secrets: %v\n", err)
				os.Exit(1)
			}
			if confirmed {
				_, err := client.SetRepository(cmd.Context(), cfg.GithubInfo.Organization, cfg.GithubInfo.Repository)
				if err != nil {
					fmt.Printf("error setting repository: %v\n", err)
					os.Exit(1)
				}

				for _, secret := range secrets {
					if err := client.SetEncryptedSecret(cmd.Context(), secret.Name, secret.Value); err != nil {
						fmt.Printf("error setting %q: %v\n", secret.Name, err)
						os.Exit(1)
					}
				}
			}

			if _, err := exec.Command("git", "remote", "add", "origin", url).CombinedOutput(); err != nil {
				fmt.Printf("error running git: %v\n", err)
				os.Exit(1)
			}

			if _, err := exec.Command("git", "push", "origin", "main").CombinedOutput(); err != nil {
				fmt.Printf("error running git: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	initCmd.Flags().BoolVarP(&skipTidy, "skip-tidy", "s", false, "Skip cleaning up the rendered output files")
	initCmd.Flags().BoolVar(&initializeRepo, "initialize-repo", false, "Initialize repo")

	rootCmd.AddCommand(initCmd)
}
