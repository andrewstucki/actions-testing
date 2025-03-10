// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package config

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
	Mappings map[string]string `yaml:"mappings,omitempty"`
	Bot      BotInfo           `yaml:"bot"`
}

type ConfigFile struct {
	License    LicenseInfo   `yaml:"license"`
	GithubInfo GithubInfo    `yaml:"github"`
	Projects   []ProjectInfo `yaml:"projects"`
	Backports  BackportInfo  `yaml:"backports"`
}
