version: '3'

# if a task is referenced multiple times, only run it once
run: once

# configure bash to recursively expand **
shopt: [globstar]

tasks:
  {{- if .LicenseManagement }}
  generate-third-party-licenses:
    dir: {{ .Source }}
    method: checksum
    generates:
      - third_party_licenses.md
    sources:
      - ./go.mod
      - ./go.sum
    cmds: 
      - |
        go-licenses report ./... --template ../support/files/third_party_licenses.md.tpl \
        --ignore {{ .GithubURL }} > ../third_party_licenses.md

  write-license-headers:
    cmds:
      - licenseupdater
  {{- end }}

  pending-prs:
    desc: "Get all pending PRs for watched branches"
    silent: true
    cmds:
      - ./.github/workflows/scripts/pending-prs terminal {{ .Organization }}/{{ .Repository }}
