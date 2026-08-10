# github.com/papercomputeco/daggerverse/dockerimage

Builds a Dockerfile locally or publishes a multi-platform OCI image.

Publishing defaults to `linux/amd64` and `linux/arm64`. Registry authentication
is supplied by the caller's Dagger environment.

Run the image fixtures for both platforms with `make test`.

## Usage

### Build locally

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/dockerimage \
  --source=. \
  build \
  export-image --name=my-app:dev
```

Build a specific platform with a build argument, without loading it:

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/dockerimage \
  --source=. \
  with-build-arg --name=LDFLAGS --value="-s -w" \
  build --platform=linux/arm64 \
  sync
```

Repeat `with-build-arg` to pass more than one argument.

### Publish

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/dockerimage \
  --source=. \
  publish \
    --repository=registry.example.com/team/my-app \
    --tags=v1.0.0 \
    --tags=latest
```

Use constructor options to select another Dockerfile or replace the default
publishing platforms.

For release workflows, pin the module to a commit instead of floating on the
default branch.
