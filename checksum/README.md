# github.com/papercomputeco/daggerverse/checksum

Recursively generates SHA256 checksums for every file in a directory.

Each input file gets a sibling `.sha256` file containing its checksum.
For example, `bin/myapp` produces `bin/myapp.sha256`.


| Function | Description |
|----------|-------------|
| `checksum` | Accepts a directory, generates a `.sha256` file for each file, and returns the directory with checksums included. |


## Usage

### Generate checksums for a build directory

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/checksum \
  checksum \
    --dir ./dist \
  export --path ./dist
```

### Pipe from another module

```sh
dagger call \
  -m github.com/papercomputeco/daggerverse/checksum \
  checksum \
    --dir ./release-artifacts \
  export --path ./release-artifacts
```
