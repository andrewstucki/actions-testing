// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package templates

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
	"text/template"
)

var (
	//go:embed all:files
	templateFiles embed.FS

	defaultBackportLabel             = "backport"
	defaultGithubBackportBot         = "github-actions[bot]"
	defaultGithubBackportBotTokenVar = "GITHUB_TOKEN"
	defaultRenderer                  = &Renderer{}
)

// TemplateInfo is the info to render into our templates.
type TemplateInfo struct {
	Copyright            string
	Organization         string
	Repository           string
	BackportBranches     []string
	Versions             []string
	Projects             []ProjectInfo
	Label                string
	LabelMapper          map[string]string
	Source               string
	License              string
	BackportBot          string
	BackportBotTokenVar  string
	LicenseManagement    bool
	Backports            bool
	AutoApproveBackports bool
}

// ProjectInfo is the info of a project with a mapping to its Changelog
type ProjectInfo struct {
	Name      string
	Changelog string
}

// NormalizeAndValidate returns an error if the TemplateInfo is invalid
func (t *TemplateInfo) NormalizeAndValidate() error {
	// Normalization
	if t.Source == "" {
		t.Source = "."
	}
	if t.Copyright == "" {
		t.Copyright = t.Organization
	}
	if t.BackportBot == "" {
		t.BackportBot = defaultGithubBackportBot
	}
	if t.BackportBotTokenVar == "" {
		t.BackportBotTokenVar = defaultGithubBackportBotTokenVar
	}
	if t.Label == "" {
		t.Label = defaultBackportLabel
	}
	if len(t.LabelMapper) == 0 {
		t.LabelMapper = map[string]string{
			"^v(\\d+).(\\d+).\\d+$": "v$1.$2.x",
		}
	}
	if len(t.Projects) == 0 {
		t.Projects = append(t.Projects, ProjectInfo{
			Name:      t.Repository,
			Changelog: "CHANGELOG.md",
		})
	}

	// Validation
	var errs []error
	if t.Organization == "" {
		errs = append(errs, errors.New("Organization must be specified"))
	}
	if t.Repository == "" {
		errs = append(errs, errors.New("Repository must be specified"))
	}
	if t.License == "" {
		errs = append(errs, errors.New("License must be specified"))
	}

	return errors.Join(errs...)
}

// GithubURL returns the github path to this Go project
func (t TemplateInfo) GithubURL() string {
	github := "github.com/"
	if t.Source != "." {
		return github + t.Organization + "/" + t.Repository + "/" + t.Source
	}
	return github + t.Organization + "/" + t.Repository
}

func (t TemplateInfo) JSONBranches() string {
	branches, err := json.Marshal(t.BackportBranches)
	if err != nil {
		panic(fmt.Errorf("creating branches: %w", err))
	}
	return string(branches)
}

func (t TemplateInfo) JSONBranchesWithMain() string {
	branches, err := json.Marshal(append([]string{"main"}, t.BackportBranches...))
	if err != nil {
		panic(fmt.Errorf("creating branches: %w", err))
	}
	return string(branches)
}

func (t TemplateInfo) JSONLabelMappings() string {
	mappings, err := json.MarshalIndent(t.LabelMapper, "", "    ")
	if err != nil {
		panic(fmt.Errorf("creating label mappings: %w", err))
	}
	return string(mappings)
}

// File is the representation of a rendered file.
type File struct {
	// Name of the rendered file.
	Name string
	// Once means that the file shouldn't be overwritten
	// after the first time it's rendered. If the file exists
	// it will not be written again.
	Once bool
	// Executable means the file should be marked executable when rendered
	Executable bool
	// The rendered file contents
	Data []byte
}

// Renderer customizes the rendering behavior of the templates.
type Renderer struct {
	// Ignore rendering a template only once, always render it
	// even if it already exists.
	IgnoreOnce bool
	// Ignore setting executable permissions on files.
	IgnoreExecutable bool
	// Suffix adds the given suffix to every file
	Suffix string
}

// Render renders templates to the filesystem using the default renderer.
func RenderTo(directory string, info TemplateInfo) error {
	return defaultRenderer.RenderTo(directory, info)
}

// RenderTo renders all of our templates using the given info
// into the given directory.
func (r *Renderer) RenderTo(directory string, info TemplateInfo) error {
	files, err := Render(info)
	if err != nil {
		return err
	}

	var errs []error
	for _, file := range files {
		fileName := path.Join(directory, file.Name)
		parent := path.Dir(fileName)
		if err := os.MkdirAll(parent, 0755); err != nil {
			errs = append(errs, fmt.Errorf("creating parent directory: %w", err))
			continue
		}

		if r.Suffix != "" {
			fileName += "." + r.Suffix
		}

		_, err := os.Stat(fileName)
		if os.IsNotExist(err) || !file.Once || r.IgnoreOnce {
			permissions := os.FileMode(0644)
			if file.Executable && !r.IgnoreExecutable {
				permissions = 0755
			}
			if err := os.WriteFile(fileName, file.Data, permissions); err != nil {
				errs = append(errs, fmt.Errorf("writing file: %w", err))
			}
		}
	}

	return errors.Join(errs...)
}

// Render renders templates to in memory files using the default renderer.
func Render(info TemplateInfo) ([]File, error) {
	return defaultRenderer.Render(info)
}

// Render renders all of our templates using the given info
// and returns all of the rendered templates.
func (r *Renderer) Render(info TemplateInfo) ([]File, error) {
	if err := info.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	var renderedFiles []File
	var errs []error

	if err := fs.WalkDir(templateFiles, ".", func(fullPath string, d fs.DirEntry, err error) error {
		if err != nil {
			errs = append(errs, err)
			return nil
		}

		if !d.IsDir() {
			var buffer bytes.Buffer
			once := false

			name := strings.TrimPrefix(fullPath, "files/")
			isTemplate := strings.HasSuffix(name, ".tpl")
			if isTemplate {
				name = strings.TrimSuffix(name, ".tpl")
				once = strings.HasSuffix(name, ".once")
				name = strings.TrimSuffix(name, ".once")
			}

			isFile := strings.HasSuffix(name, ".file")
			if isFile {
				once = true
				name = strings.TrimSuffix(name, ".file")
			}

			if !isFile && !isTemplate {
				return nil
			}

			isExecute := strings.HasSuffix(name, ".execute")
			name = strings.TrimSuffix(name, ".execute")

			data, err := templateFiles.ReadFile(fullPath)
			if err != nil {
				errs = append(errs, err)
				return nil
			}

			switch {
			case isTemplate:
				tmpl, err := template.New("").Parse(string(data))
				if err != nil {
					errs = append(errs, err)
					return nil
				}

				err = tmpl.Execute(&buffer, info)
				if err != nil {
					errs = append(errs, err)
					return nil
				}
			case isFile:
				if _, err := buffer.WriteString(string(data)); err != nil {
					errs = append(errs, err)
					return nil
				}
			default:
				errs = append(errs, fmt.Errorf("unknown template type for file: %q", fullPath))
				return nil
			}

			// don't render any conditionally rendered files
			if strings.TrimSpace(buffer.String()) == "" {
				return nil
			}

			renderedFiles = append(renderedFiles, File{
				Data:       buffer.Bytes(),
				Name:       name,
				Once:       once,
				Executable: isExecute,
			})
		}

		return nil
	}); err != nil {
		errs = append(errs, err)
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return renderedFiles, nil
}
