{{- if .LicenseManagement -}}
organization: {{ .Copyright }}
top_level_license: {{ .License }}
matches:
  - type: go
    short: true
    extension: .go
    directory: {{ .Source }}
    license: {{ .License }}
{{- end -}}