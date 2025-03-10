changesDir: .changes
unreleasedDir: unreleased
headerPath: header.tpl.md
changelogPath: CHANGELOG.md
versionExt: md
versionFormat: '## {{ "{{.Version}} - {{.Time.Format \"2006-01-02\"}}" }}'
kindFormat: '### {{ "{{.Kind}}" }}'
changeFormat: '* {{ "{{.Body}}" }}'
body:
  block: true
# All changes specify auto as 'patch' to avoid unintentional major or minor
# version bumps as those are handled manually.
kinds:
    - label: Added
      auto: patch
    - label: Changed
      auto: patch
    - label: Deprecated
      auto: patch
    - label: Removed
      auto: patch
    - label: Fixed
      auto: patch
newlines:
    afterChangelogHeader: 1
    beforeChangelogVersion: 1
    endOfVersion: 1
envPrefix: CHANGIE_
# Project keys and version separators are configured to align with the tagging
# semantics of multi-module repositories. `dir/of/module/v<version>`
# https://go.dev/wiki/Modules#what-are-multi-module-repositories
projectsVersionSeparator: "/"
projects:
{{- range $project := .Projects }}
- label: {{ $project.Name }}
  key: {{ $project.Name }}
  changelog: {{ $project.Changelog }}
{{- end }}