// Utils are the catch all, useful for various other modules
package main

import (
	"context"
	"dagger/utils/internal/dagger"
	"fmt"
	"strings"
)

type Utils struct{}

// FlattenNameOsArch takes a build artifact directory organized as <os>/<arch>/<filename>
// and returns a flat directory with files renamed to <filename>-<os>-<arch>
// (or <filename>-<os>-<arch>.sha256 for checksum files).
// This is a standalone utility — for the chained workflow, use WithFlatten instead.
func (m *Utils) FlattenNameOsArch(
	ctx context.Context,

	// Directory containing build artifacts organized as <os>/<arch>/<filename>
	build *dagger.Directory,
) (*dagger.Directory, error) {
	entries, err := build.Glob(ctx, "*/*/*")
	if err != nil {
		return nil, fmt.Errorf("failed to list build artifacts: %w", err)
	}

	dist := dag.Directory()

	for _, entry := range entries {
		parts := strings.SplitN(entry, "/", 3)
		if len(parts) != 3 {
			continue
		}
		os := parts[0]
		arch := parts[1]
		filename := parts[2]

		var newName string
		if strings.HasSuffix(filename, ".sha256") {
			base := strings.TrimSuffix(filename, ".sha256")
			newName = fmt.Sprintf("%s-%s-%s.sha256", base, os, arch)
		} else {
			newName = fmt.Sprintf("%s-%s-%s", filename, os, arch)
		}

		dist = dist.WithFile(newName, build.File(entry))
	}

	return dist, nil
}
