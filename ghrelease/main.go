// GitHub release management.
//
// Upload takes a directory of build artifacts to a release.
// Artifacts can optionally be flattened from an
// <os>/<arch>/<filename> layout into a flat directory with files renamed
// to <filename>-<os>-<arch> (with .sha256 checksum extensions preserved).
// For example "darwin/arm64/tapes" and "darwin/arm64/tapes.sha256" can be flattened
// to "tapes-darwin-arm64" and "tapes-darwin-arm64.sha256".

package main

import (
	"context"
	"fmt"
	"path"

	"dagger/ghrelease/internal/dagger"
)

// Ghrelease manages github releases.
type Ghrelease struct {
	// GitHub token for authentication
	//
	// +private
	Token *dagger.Secret

	// GitHub repository in owner/repo format (e.g., "papercomputeco/tapes")
	//
	// +private
	Repo string

	// Directory of assets to upload
	//
	// +private
	Assets *dagger.Directory

	// Whether to flatten Assets from <os>/<arch>/<filename> layout before uploading
	//
	// +private
	DoFlatten bool

	// Release tag for the upload
	//
	// +private
	Tag string
}

// New creates a new Ghrelease instance.
func New(
	// GitHub token with permissions to upload release assets
	token *dagger.Secret,

	// GitHub repository in owner/repo format (e.g., "papercomputeco/myproject")
	repo string,

	// Directory of assets to upload
	assets *dagger.Directory,
) *Ghrelease {
	return &Ghrelease{
		Token:  token,
		Repo:   repo,
		Assets: assets,
	}
}

// WithFlatten enables flattening of the assets directory before upload.
// When chained, the <os>/<arch>/<filename> directory structure is collapsed
// into a flat directory with files renamed to <filename>-<os>-<arch>
// (or <filename>-<os>-<arch>.sha256 for checksum files).
func (m *Ghrelease) WithFlatten() *Ghrelease {
	m.DoFlatten = true
	return m
}

// WithTag stores the release tag for upload.
func (m *Ghrelease) WithTag(
	// Release tag to upload assets to (e.g., "nightly", "v1.0.0")
	tag string,
) *Ghrelease {
	m.Tag = tag
	return m
}

// Upload uploads all assets to a GitHub release.
// If WithFlatten was chained, the assets are flattened first.
// The tag must have been set via WithTag before calling Upload.
func (m *Ghrelease) Upload(ctx context.Context) error {
	if m.Tag == "" {
		return fmt.Errorf("no tag set: call WithTag before Upload")
	}

	dist := m.Assets

	if m.DoFlatten {
		dist = dag.Utilsverse().FlattenNameOsArch(m.Assets)
	}

	entries, err := dist.Glob(ctx, "*")
	if err != nil {
		return fmt.Errorf("failed to list dist files: %w", err)
	}

	uploadArgs := []string{
		"gh", "release", "upload", m.Tag,
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
