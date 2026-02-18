// Upload build artifacts to GitHub releases.
//
// This module takes a directory of build artifacts organized by OS and architecture
// (e.g., <os>/<arch>/<filename>) and uploads them to a GitHub release.
// Files are renamed to include the OS and architecture in their name
// (e.g., <filename>-<os>-<arch>), and .sha256 checksum files are handled
// so the extension stays at the end (e.g., <filename>-<os>-<arch>.sha256).

package main

import (
	"context"
	"fmt"
	"path"
	"strings"

	"dagger/ghrelease/internal/dagger"
)

// Ghrelease uploads build artifacts to GitHub releases.
type Ghrelease struct {
	// GitHub token for authentication
	//
	// +private
	Token *dagger.Secret

	// GitHub repository in owner/repo format (e.g., "papercomputeco/myproject")
	//
	// +private
	Repo string
}

// New creates a new Ghrelease instance.
func New(
	// GitHub token with permissions to upload release assets
	token *dagger.Secret,

	// GitHub repository in owner/repo format (e.g., "papercomputeco/myproject")
	repo string,
) *Ghrelease {
	return &Ghrelease{
		Token: token,
		Repo:  repo,
	}
}

// Flatten takes a build artifact directory organized as <os>/<arch>/<filename>
// and returns a flat directory with files renamed to <filename>-<os>-<arch>
// (or <filename>-<os>-<arch>.sha256 for checksum files).
func (m *Ghrelease) Flatten(
	ctx context.Context,

	// Directory containing build artifacts organized as <os>/<arch>/<filename>
	build *dagger.Directory,
) (*dagger.Directory, error) {
	// Glob all files at the expected depth: <os>/<arch>/<filename>
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

// Upload uploads all files in the given directory to a GitHub release.
// The directory should be flat (no subdirectories) — use Flatten first
// if you need to rename build artifacts from an <os>/<arch>/<filename> layout.
func (m *Ghrelease) Upload(
	ctx context.Context,

	// Directory containing files to upload as release assets
	dist *dagger.Directory,

	// Release tag to upload assets to (e.g., "nightly", "v1.0.0")
	tag string,
) error {
	entries, err := dist.Glob(ctx, "*")
	if err != nil {
		return fmt.Errorf("failed to list dist files: %w", err)
	}

	uploadArgs := []string{
		"gh", "release", "upload", tag,
		"--repo", m.Repo,
		"--clobber",
	}
	for _, entry := range entries {
		uploadArgs = append(uploadArgs, path.Join("/dist", entry))
	}

	_, err = dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "github-cli"}).
		WithSecretVariable("GH_TOKEN", m.Token).
		WithDirectory("/dist", dist).
		WithExec(uploadArgs).
		Sync(ctx)

	if err != nil {
		return fmt.Errorf("failed to upload release assets: %w", err)
	}

	return nil
}
