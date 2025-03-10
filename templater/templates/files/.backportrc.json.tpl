{{- if .Backports -}}
{
    "fork": false,
    "repoOwner": "{{ .Organization }}",
    "repoName": "{{ .Repository }}",
    "autoMerge": true,
    "targetBranchChoices": {{ .JSONBranches }},
    "targetPRLabels": ["{{ .Label }}"],
    "branchLabelMapping": {{ .JSONLabelMappings }}
}
{{- end -}}
