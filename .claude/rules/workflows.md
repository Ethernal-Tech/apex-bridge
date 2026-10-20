---
paths:
  - ".github/workflows/**"
  - ".github/actions/**"
---

# GitHub Actions workflows

- Pin every action by commit SHA with the tag in a comment
  (`uses: actions/checkout@<sha> # v5.0.0`). Why: a tag can be moved to other code.
- Set `permissions: {}` at the top and grant each job only what it needs. Why: a compromised
  step can do only what its token allows.
- Pass `${{ github.event.* }}` and other untrusted input to `run:` through `env:`, never inline.
  Why: an inline expression is pasted into the script before it runs (script injection).
- Give every job a `timeout-minutes`. Why: a hung job burns minutes until the 6-hour limit.
- Run `actionlint` and `zizmor` on a changed workflow before the PR.
