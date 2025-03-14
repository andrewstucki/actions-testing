name: 'Notify of Pending PRs'

on:
  workflow_dispatch:

jobs:
  stale:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Generate Pending PRs list
        id: generate-prs
        run: |
          ./.github/workflows/scripts/pending-prs slack {{ .Organization }}/{{ .Repository }} > payload.json
          echo "has-prs=$(cat payload.json | wc -l)" >> $GITHUB_OUTPUT
        env:
          GH_TOKEN: {{ "${{ github.token }}"}}
      - name: Post message to Slack channel
        uses: slackapi/slack-github-action@v2.0.0
        if: steps.generate-prs.outputs.has-prs != '0'
        with:
          webhook: {{ "${{ secrets.SLACK_WEBHOOK_URL }}"}}
          webhook-type: webhook-trigger
          payload-file-path: "./payload.json"