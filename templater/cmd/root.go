// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/andrewstucki/actions-testing/templater/templates"
)

var configFile string

type LicenseInfo struct {
	Copyright string `yaml:"copyright"`
	License   string `yaml:"license"`
}

type ProjectInfo struct {
	Name      string `yaml:"name"`
	Changelog string `yaml:"changelog"`
}

type GithubInfo struct {
	Organization string `yaml:"organization"`
	Repository   string `yaml:"repository"`
}

type BotInfo struct {
	Name          string `yaml:"name"`
	TokenVariable string `yaml:"token_variable"`
}

type BackportInfo struct {
	Label    string            `yaml:"label"`
	Branches []string          `yaml:"branches"`
	Versions []string          `yaml:"versions"`
	Mappings map[string]string `yaml:"mappings"`
	Bot      BotInfo           `yaml:"bot"`
}

type ConfigFile struct {
	License    LicenseInfo   `yaml:"license"`
	GithubInfo GithubInfo    `yaml:"github"`
	Projects   []ProjectInfo `yaml:"projects"`
	Backports  BackportInfo  `yaml:"backports"`
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "templater",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(configFile)
		if err != nil {
			fmt.Printf("error reading configuration file: %v\n", err)
			os.Exit(1)
		}

		var config ConfigFile
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			fmt.Printf("error unmarshaling configuration file: %v\n", err)
			os.Exit(1)
		}

		info := templates.TemplateInfo{
			Copyright:            config.License.Copyright,
			License:              config.License.License,
			Organization:         config.GithubInfo.Organization,
			Repository:           config.GithubInfo.Repository,
			BackportBranches:     config.Backports.Branches,
			Versions:             config.Backports.Versions,
			Label:                config.Backports.Label,
			LabelMapper:          config.Backports.Mappings,
			BackportBot:          config.Backports.Bot.Name,
			BackportBotTokenVar:  config.Backports.Bot.TokenVariable,
			LicenseManagement:    true,
			Backports:            true,
			AutoApproveBackports: true,
		}

		for _, project := range config.Projects {
			info.Projects = append(info.Projects, templates.ProjectInfo{
				Name:      project.Name,
				Changelog: project.Changelog,
			})
		}

		if err := templates.Update.RenderTo(".", info); err != nil {
			fmt.Printf("error reading templates: %v\n", err)
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
	rootCmd.Flags().StringVarP(&configFile, "config", "c", ".template.yaml", "Location for the template configuration file.")
}
