package main

import (
	"dagger/checksum/internal/dagger"
)

type Checksumer struct{}

// Checksum recursively generates SHA256 checksums for all files in the given directory.
// These files land as `/some/path/filename.sha256`, `/some/other/path/filename.sha256`.
func (m *Checksumer) Checksum(dir *dagger.Directory) *dagger.Directory {
	return dag.Container().
		From("alpine:latest").
		WithDirectory("/artifacts", dir).
		WithWorkdir("/artifacts").
		WithExec([]string{"sh", "-c", `
			find . -type f ! -name "*.sha256" | while read file; do
				sha256sum "$file" | sed 's|./||' > "${file}.sha256"
			done
		`}).
		Directory("/artifacts")
}
