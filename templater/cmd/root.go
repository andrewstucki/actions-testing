// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrewstucki/actions-testing/templater/config"
	"github.com/andrewstucki/actions-testing/templater/templates"
)

var configFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "templater",
	Short: "A brief description of your application",
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

		if err := templates.Update.RenderTo(".", info); err != nil {
			fmt.Printf("error rendering templates: %v\n", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", ".template.yaml", "Location for the template configuration file.")
}
