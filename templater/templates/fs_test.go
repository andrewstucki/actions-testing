// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package templates

import (
	"io/fs"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const goldenFileDirectory = "testdata"

var (
	renderGolden   = false
	goldenRenderer = &Renderer{
		IgnoreOnce:       true,
		IgnoreExecutable: true,
		Suffix:           "golden",
	}
)

func init() {
	if os.Getenv("RENDER_GOLDEN_FILES") == "true" {
		renderGolden = true
	}
}

func TestRenderTo(t *testing.T) {
	directory := t.TempDir()
	t.Cleanup(func() {
		if err := os.RemoveAll(directory); err != nil {
			t.Logf("error removing directory %q: %v", directory, err)
		}
	})

	for name, tt := range map[string]struct {
		info TemplateInfo
		err  error
	}{
		"basic": {
			info: TemplateInfo{
				Organization: "org",
				Repository:   "repo",
				License:      "MIT",
			},
		},
		"license": {
			info: TemplateInfo{
				LicenseManagement: true,
				Organization:      "org",
				Repository:        "repo",
				License:           "MIT",
			},
		},
		"source": {
			info: TemplateInfo{
				Source:            "source",
				LicenseManagement: true,
				Organization:      "org",
				Copyright:         "My Organization",
				Repository:        "repo",
				License:           "MIT",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			fileName := strings.SplitN(t.Name(), "/", 2)[1]
			goldenDirectory := path.Join(goldenFileDirectory, fileName)
			testDirectory := path.Join(directory, fileName)

			if renderGolden && tt.err == nil {
				require.NoError(t, goldenRenderer.RenderTo(goldenDirectory, tt.info))
			}

			if tt.err != nil {
				require.EqualError(t, RenderTo(testDirectory, tt.info), tt.err.Error())
			} else {
				require.NoError(t, RenderTo(testDirectory, tt.info))
				requireGoldenDirectoriesEqual(t, goldenDirectory, testDirectory)
			}
		})

	}
}

type testFile struct {
	name string
	data string
}

func requireGoldenDirectoriesEqual(t *testing.T, expected, actual string) {
	t.Helper()

	mapToFiles := func(directory, suffix string) []testFile {
		mapped := []testFile{}
		dirFS := os.DirFS(directory)
		err := fs.WalkDir(dirFS, ".", func(fullPath string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			data, err := fs.ReadFile(dirFS, fullPath)
			if err != nil {
				return err
			}

			mapped = append(mapped, testFile{
				name: strings.TrimSuffix(d.Name(), "."+suffix),
				data: string(data),
			})
			return nil
		})
		require.NoError(t, err)
		return mapped
	}

	expectedFiles := mapToFiles(expected, "golden")
	actualFiles := mapToFiles(actual, "")

	require.ElementsMatch(t, expectedFiles, actualFiles)
}
