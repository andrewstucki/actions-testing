// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package prompt

import (
	"fmt"
	"strings"

	"github.com/cqroot/prompt"
	"github.com/cqroot/prompt/input"

	"github.com/andrewstucki/actions-testing/templater/config"
)

type configFunc func(cfg *config.ConfigFile, field string) error

func wrapConfigFunc(cfg *config.ConfigFile, cfgFn configFunc) input.ValidateFunc {
	return func(value string) error {
		return cfgFn(cfg, value)
	}
}

func setFieldNotEmpty(fieldValue *string, fieldDescription, field string) error {
	field = strings.TrimSpace(field)
	if field == "" {
		return fmt.Errorf("%s is required", fieldDescription)
	}
	*fieldValue = field
	return nil
}

func setFieldDefault(fieldValue *string, defaultValue, field string) error {
	field = strings.TrimSpace(field)
	if field == "" {
		*fieldValue = defaultValue
	} else {
		*fieldValue = field
	}
	return nil
}

type field struct {
	prompt         string
	defaultValue   string
	defaultValueFn func(cfg *config.ConfigFile) string
	isNumber       bool
	validator      configFunc
}

// package config

// type LicenseInfo struct {
// 	Copyright string `yaml:"copyright"`
// 	License   string `yaml:"license"`
// }

// type ProjectInfo struct {
// 	Name      string `yaml:"name"`
// 	Changelog string `yaml:"changelog"`
// }

// type GithubInfo struct {
// 	Organization string `yaml:"organization"`
// 	Repository   string `yaml:"repository"`
// }

// type BotInfo struct {
// 	Name          string `yaml:"name"`
// 	TokenVariable string `yaml:"token_variable"`
// }

// type BackportInfo struct {
// 	Label    string            `yaml:"label"`
// 	Branches []string          `yaml:"branches"`
// 	Versions []string          `yaml:"versions"`
// 	Mappings map[string]string `yaml:"mappings"`
// 	Bot      BotInfo           `yaml:"bot"`
// }

// type ConfigFile struct {
// 	License    LicenseInfo   `yaml:"license"`
// 	GithubInfo GithubInfo    `yaml:"github"`
// 	Projects   []ProjectInfo `yaml:"projects"`
// 	Backports  BackportInfo  `yaml:"backports"`
// }

var prompts = []field{
	{prompt: "Name of your project", validator: func(cfg *config.ConfigFile, field string) error {
		return setFieldNotEmpty(&cfg.GithubInfo.Repository, "project name", field)
	}},
	{prompt: "Github organization", validator: func(cfg *config.ConfigFile, field string) error {
		return setFieldNotEmpty(&cfg.GithubInfo.Organization, "organization name", field)
	}},
	{prompt: "License", defaultValue: "MIT", validator: func(cfg *config.ConfigFile, field string) error {
		return setFieldDefault(&cfg.License.License, "MIT", field)
	}},
	{prompt: "Copyright holder", defaultValueFn: func(cfg *config.ConfigFile) string { return cfg.GithubInfo.Organization }, validator: func(cfg *config.ConfigFile, field string) error {
		return setFieldDefault(&cfg.License.Copyright, cfg.GithubInfo.Organization, field)
	}},
	{prompt: "Github backport user", defaultValue: "github-actions[bot]", validator: func(cfg *config.ConfigFile, field string) error {
		return setFieldDefault(&cfg.Backports.Bot.Name, "github-actions[bot]", field)
	}},
	{prompt: "Github backport token variable", defaultValue: "GITHUB_TOKEN", validator: func(cfg *config.ConfigFile, field string) error {
		return setFieldDefault(&cfg.Backports.Bot.TokenVariable, "GITHUB_TOKEN", field)
	}},
}

func Run() (*config.ConfigFile, error) {
	cfg := &config.ConfigFile{}
	prompter := prompt.New()

	for _, field := range prompts {
		opts := []input.Option{
			input.WithHelp(true),
			input.WithValidateFunc(wrapConfigFunc(cfg, field.validator)),
		}

		if field.isNumber {
			opts = append(opts, input.WithInputMode(input.InputNumber))
		}

		defaultValue := field.defaultValue
		if field.defaultValueFn != nil {
			defaultValue = field.defaultValueFn(cfg)
		}

		_, err := prompter.Ask(field.prompt).Input(defaultValue, opts...)
		if err != nil {
			return nil, err
		}
	}

	// set the rest of our defaults
	cfg.Backports.Branches = []string{"v0.0.x"}
	cfg.Backports.Versions = []string{"v0.0.1"}
	cfg.Backports.Label = "backport"
	cfg.Backports.Mappings = map[string]string{
		"^v(\\d+).(\\d+).\\d+$": "v$1.$2.x",
	}
	cfg.Projects = append(cfg.Projects, config.ProjectInfo{
		Name:      cfg.GithubInfo.Repository,
		Changelog: "CHANGELOG.md",
	})

	return cfg, nil
}
