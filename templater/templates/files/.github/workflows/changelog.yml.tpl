name: Changelog

on:
  pull_request:
    branches:
      # only check for changelog entries going into main 
      - main

jobs:
  changed_files:
    if: {{ "${{ !contains(github.event.pull_request.labels.*.name, 'no-changelog') }}" }}
    runs-on: ubuntu-latest
    name: Check for changelog entry
    steps:
      - uses: actions/checkout@v4

      - name: Get all changed changelog files
        id: changed-changelog-files
        uses: tj-actions/changed-files@v45
        with:
          files: |
            .changes/unreleased/**.yaml

      - name: Pass
        if: steps.changed-changelog-files.outputs.any_changed == 'true'
        run: |
          echo "Found changelog entry"

      - name: Fail
        if: steps.changed-changelog-files.outputs.any_changed != 'true'
        run: |
          echo "No changelog entry detected." && exit 1
