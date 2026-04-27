# github.com/papercomputeco/daggerverse/ghcontrib

Various GitHub contribution utilities and standards conforming checks.

<<<<<<< add-pr-linear-checker
Currently includes checks for:
- required conventional PR title prefixes:
  - `✨ feat: ` — new features
  - `🔧 fix: ` — bug fixes
  - `🧹 chore: ` — maintenance tasks
  - `♻️ refactor: ` — refactors
- Linear magic-word references in the PR title or body, e.g. `fixes PCC-123` or `related to DES-456`
=======
Currently checks that PR titles start with a required conventional prefix:
- `✨ feat: ` — new features
- `🔧 fix: ` — bug fixes
- `🧹 chore: ` — maintenance tasks
- `♻️ refactor: ` — code restructuring without behavior changes
- `🎨 design: ` — design and UX updates
>>>>>>> main

Supported Linear teams are `PCC`, `DES`, and `REL`.

| Function | Description |
|----------|-------------|
| `check-pull-request` | Fetches a PR via `gh pr view` and validates that its title starts with a required prefix. Fails the pipeline if the PR is non-conforming. |
| `check-pull-request-linear-magic-word` | Fetches a PR via `gh pr view` and validates only the Linear magic-word reference. |


## Constructor arguments

| Argument | Type | Description |
|----------|------|-------------|
| `--token` | `Secret` | GitHub token for authenticating with the `gh` CLI |
| `--repo` | `String` | GitHub repository in `owner/repo` format |


## Usage

### Check a pull request in CI

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghcontrib \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  check-pull-request --number 42
```

### Check only the Linear magic-word reference

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghcontrib \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  check-pull-request-linear-magic-word --number 42
```

### From a GitHub Actions workflow

```yaml
- name: Check PR title
  uses: dagger/dagger-for-github@v7
  with:
    version: "latest"
    module: github.com/papercomputeco/daggerverse/ghcontrib
    verb: call
    args: >-
      --token env:GITHUB_TOKEN
      --repo "${{ github.repository }}"
      check-pull-request --number "${{ github.event.pull_request.number }}"
```
