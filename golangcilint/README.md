# github.com/papercomputeco/daggerverse/golangcilint

A cacheable, containerized Dagger module that runs [golangci-lint](https://golangci-lint.run/)
against Go source code.


| Function | Description |
|----------|-------------|
| `check`  | Runs golangci-lint and returns the output. Fails the pipeline on lint violations (ideal for CI). |
| `lint`   | Runs golangci-lint with `--fix` and returns the source directory with auto-fixes applied. |


## Default config

When no `--config` is provided, the module applies a minimal built-in
`.golangci.yml` that uses the standard linter set, uncaps issue limits, and
sets a 5 minute timeout. Most projects should provide their own config file
via `--config` for project-specific rules.


## Usage

### Run a lint check in CI

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/golangcilint \
  --source . \
  check
```

### Auto-fix lint issues and export the result

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/golangcilint \
  --source . \
  lint \
  export --path .
```

### Provide a custom config file

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/golangcilint \
  --source . \
  --config .golangci.yml \
  check
```
