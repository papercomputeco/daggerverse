# Go toolchain

Provides Go toolchain checks and build utils.

Install with:

```
dagger toolchain install github.com/papercomputeco/daggerverse/go
```

Run:

```
dagger check -l
```

## Environment variables

Pass `--env-vars` to set environment variables (in `KEY=VALUE` format) in the
Go container, e.g. for `GOEXPERIMENT` flags. As a toolchain, set it via a
customization in `dagger.json`:

```json
{
  "toolchains": [
    {
      "name": "go",
      "source": "github.com/papercomputeco/daggerverse/go@main",
      "customizations": [
        {
          "argument": "envVars",
          "default": "GOEXPERIMENT=jsonv2"
        }
      ]
    }
  ]
}
```
