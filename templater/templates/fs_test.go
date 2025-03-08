// Copyright (c) Andrew Stucki
// SPDX-License-Identifier: MIT

package templates

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

const goldenFileDirectory = "testdata"

var renderGolden = false

func init() {
	ignoreOnce = true
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
			},
		},
		"license": {
			info: TemplateInfo{
				LicenseManagement: true,
				Organization:      "org",
				Repository:        "repo",
			},
		},
		"source": {
			info: TemplateInfo{
				Source:            "source",
				LicenseManagement: true,
				Organization:      "org",
				Repository:        "repo",
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			fileName := strings.SplitN(t.Name(), "/", 2)[1]
			goldenDirectory := path.Join(goldenFileDirectory, fileName)
			testDirectory := path.Join(directory, fileName)

			if renderGolden && tt.err == nil {
				require.NoError(t, RenderTo(goldenDirectory, tt.info))
			}

			if tt.err != nil {
				require.EqualError(t, RenderTo(testDirectory, tt.info), tt.err.Error())
			} else {
				require.NoError(t, RenderTo(testDirectory, tt.info))
				requireDirectoriesEqual(t, goldenDirectory, testDirectory)
			}
		})

	}
}

type testFile struct {
	name string
	data string
}

func requireDirectoriesEqual(t *testing.T, expected, actual string) {
	t.Helper()

	mapToFiles := func(directory string) []testFile {
		files, err := os.ReadDir(directory)
		require.NoError(t, err)

		mapped := []testFile{}
		for _, file := range files {
			data, err := os.ReadFile(path.Join(directory, file.Name()))
			require.NoError(t, err)
			mapped = append(mapped, testFile{
				name: file.Name(),
				data: string(data),
			})
		}
		return mapped
	}

	expectedFiles := mapToFiles(expected)
	actualFiles := mapToFiles(actual)

	require.ElementsMatch(t, expectedFiles, actualFiles)
}
