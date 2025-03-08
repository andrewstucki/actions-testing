// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package templates

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/template"
)

//go:embed files/*.tpl
var templateFiles embed.FS

var ignoreOnce = false

// TemplateInfo is the info to render into our templates.
type TemplateInfo struct {
	Organization      string
	Repository        string
	Versions          []string
	Source            string
	LicenseManagement bool
}

// NormalizeAndValidate returns an error if the TemplateInfo is invalid
func (t *TemplateInfo) NormalizeAndValidate() error {
	// Normalization
	if t.Source == "" {
		t.Source = "."
	}

	// Validation
	var errs []error
	if t.Organization == "" {
		errs = append(errs, errors.New("Organization must be specified"))
	}
	if t.Repository == "" {
		errs = append(errs, errors.New("Repository must be specified"))
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

// File is the representation of a rendered file.
type File struct {
	// Name of the rendered file.
	Name string
	// Once means that the file shouldn't be overwritten
	// after the first time it's rendered. If the file exists
	// it will not be written again.
	Once bool
	// The rendered file contents
	Data []byte
}

// RenderTo renders all of our templates using the given info
// into the given directory.
func RenderTo(directory string, info TemplateInfo) error {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}

	files, err := Render(info)
	if err != nil {
		return err
	}

	var errs []error
	for _, file := range files {
		fileName := path.Join(directory, file.Name)
		_, err := os.Stat(fileName)
		if os.IsNotExist(err) || !file.Once || ignoreOnce {
			if err := os.WriteFile(fileName, file.Data, 0644); err != nil {
				errs = append(errs, fmt.Errorf("writing file: %w", err))
			}
		}
	}

	return errors.Join(errs...)
}

// Render renders all of our templates using the given info
// and returns all of the rendered templates.
func Render(info TemplateInfo) ([]File, error) {
	if err := info.NormalizeAndValidate(); err != nil {
		return nil, err
	}

	files, err := templateFiles.ReadDir("files")
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	var errs []error
	var renderedFiles []File

	for _, file := range files {
		buffer.Reset()

		name := strings.TrimSuffix(file.Name(), ".tpl")
		once := strings.HasSuffix(name, ".once")
		name = strings.TrimSuffix(name, ".once")

		data, err := templateFiles.ReadFile("files/" + file.Name())
		if err != nil {
			errs = append(errs, err)
			continue
		}

		tmpl, err := template.New("").Parse(string(data))
		if err != nil {
			errs = append(errs, err)
			continue
		}

		err = tmpl.Execute(&buffer, info)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		renderedFiles = append(renderedFiles, File{
			Data: buffer.Bytes(),
			Name: name,
			Once: once,
		})
	}

	if len(errs) != 0 {
		return nil, errors.Join(errs...)
	}

	return renderedFiles, nil
}
