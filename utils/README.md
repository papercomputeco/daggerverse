# github.com/papercomputeco/daggerverse/utils

Catch-all utility functions useful across other modules.

Currently provides a helper for flattening build artifact directories
organized by OS and architecture into a single flat directory with
descriptive filenames.


| Function | Description |
|----------|-------------|
| `flatten-name-os-arch` | Takes an `<os>/<arch>/<filename>` directory and returns a flat directory with files renamed to `<filename>-<os>-<arch>`. Checksum files keep their `.sha256` extension (e.g., `<filename>-<os>-<arch>.sha256`). |


## Usage

### Flatten a build directory

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/utils \
  flatten-name-os-arch \
    --build ./build \
  export --path ./dist
```
