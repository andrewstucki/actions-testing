labels:
  {{- range $version := .Versions }}
  "{{ $version }}":
    color: "ededed"
  {{- end }}
  "no-changelog":
    color: "8f1402"
  "stale":
    color: "8f1402"
