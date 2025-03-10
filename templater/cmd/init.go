// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrewstucki/actions-testing/templater/prompt"
	"github.com/andrewstucki/actions-testing/templater/templates"
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
			Label:                cfg.Backports.Label,
			LabelMapper:          cfg.Backports.Mappings,
			BackportBot:          cfg.Backports.Bot.Name,
			BackportBotTokenVar:  cfg.Backports.Bot.TokenVariable,
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
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
