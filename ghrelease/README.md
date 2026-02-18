# github.com/papercomputeco/daggerverse/ghrelease

Upload build artifacts to GitHub releases.

Takes a directory of build artifacts organized by OS and architecture
(e.g., `<os>/<arch>/<filename>`), flattens them into renamed files
(e.g., `<filename>-<os>-<arch>`), and uploads them to a GitHub release.
Checksum files keep their `.sha256` extension at the end
(e.g., `<filename>-<os>-<arch>.sha256`).


| Function | Description |
|----------|-------------|
| `flatten` | Takes an `<os>/<arch>/<filename>` directory and returns a flat directory with files renamed to `<filename>-<os>-<arch>`. |
| `upload` | Uploads all files in a flat directory to a GitHub release. |


## Constructor arguments

| Argument | Type | Description |
|----------|------|-------------|
| `--token` | `Secret` | GitHub token with permissions to upload release assets |
| `--repo` | `String` | GitHub repository in `owner/repo` format |


## Usage

### Flatten and upload in one pipeline

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghrelease \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  flatten --build ./build \
  upload --tag "nightly"
```

### Flatten only (inspect or export the result)

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghrelease \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  flatten --build ./build \
  export --path ./dist
```

### Upload a pre-built directory

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/ghrelease \
  --token env:GITHUB_TOKEN \
  --repo "papercomputeco/myproject" \
  upload --dist ./dist --tag "v1.0.0"
```
