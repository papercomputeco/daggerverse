# github.com/papercomputeco/daggerverse/ghrelease

GitHub release management.

Create inspects commit subjects since the latest `v*` tag to choose the next
semantic version and generate release notes. Scoped conventional subjects are
classified the same as their unscoped form:

| Commit subject | Release behavior |
|----------------|------------------|
| `✨ feat: ...` or `✨ feat(scope): ...` | Minor bump, Features section |
| `🔧 fix: ...` or `🔧 fix(scope): ...` | Patch bump, Fixes section |
| `🧹 chore: ...` or `🧹 chore(scope): ...` | Patch bump, Chores section |
| `♻️ refactor: ...` or `♻️ refactor(scope): ...` | Patch bump, Refactors section |
| `feat!: ...`, `fix!: ...`, `chore!: ...`, or `refactor!: ...` | Major bump, Breaking Changes section |

Upload takes a directory of build artifacts to a release.
Artifacts can optionally be flattened from an
`<os>/<arch>/<filename>` layout into a flat directory with files renamed
to `<filename>-<os>-<arch>` (with `.sha256` checksum extensions preserved).
For example `darwin/arm64/tapes` and `darwin/arm64/tapes.sha256` become
`tapes-darwin-arm64` and `tapes-darwin-arm64.sha256`.


| Function | Description |
|----------|-------------|
| `with-flatten` | Enables flattening of the assets directory from `<os>/<arch>/<filename>` into `<filename>-<os>-<arch>` before upload. |
| `with-tag` | Sets the release tag for upload. |
| `upload` | Uploads all assets to a GitHub release. If `with-flatten` was chained, assets are flattened first. |


## Constructor arguments

| Argument | Type | Description |
|----------|------|-------------|
| `--token` | `Secret` | GitHub token with permissions to upload release assets |
| `--repo` | `String` | GitHub repository in `owner/repo` format |
| `--assets` | `Directory` | Directory of assets to upload |


## Usage

### Flatten and upload in one pipeline

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghrelease \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  --assets ./build \
  with-flatten \
  with-tag --tag "nightly" \
  upload
```

### Upload a pre-built flat directory

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghrelease \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  --assets ./dist \
  with-tag --tag "v1.0.0" \
  upload
```
