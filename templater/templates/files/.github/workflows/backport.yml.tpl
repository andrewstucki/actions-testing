name: Backport

on:
  pull_request_target:
    types: ["labeled", "closed"]

jobs:
  backport:
    name: Backport PR
    if: github.event.pull_request.merged == true && !(contains(github.event.pull_request.labels.*.name, 'backport'))
    runs-on: ubuntu-latest
    steps:
      - name: Backport Action
        uses: sorenlouv/backport-github-action@v9.5.1
        with:
          github_token: {{ "${{ secrets.ROBOTURTLE_TOKEN }}"}}

      - name: Info log
        if: {{ "${{ success() }}"}}
        run: cat ~/.backport/backport.info.log
        
      - name: Debug log
        if: {{ "${{ failure() }}"}}
        run: cat ~/.backport/backport.debug.log 