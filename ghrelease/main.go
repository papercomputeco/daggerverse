// GitHub release management.
//
// Create looks at git history to determine the next semver tag, generates
// categorized release notes, and creates a GitHub release via the gh CLI.

package main

import (
	"context"
	_ "embed"
	"fmt"
	"path"

	"dagger/ghrelease/internal/dagger"
)

//go:embed create-release.sh
var createReleaseScript string

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

	// Git source directory for determining release versions
	//
	// +private
	Source *dagger.Directory

	// Directory of assets to upload
	//
	// +private
	Assets *dagger.Directory

	// Whether to flatten Assets from <os>/<arch>/<filename> layout before uploading
	//
	// +private
	FlattenAssets bool

	// Release tag for the assets upload
	//
	// +private
	Tag string

	// Dry run for release creation logic. Skips the actual `gh release create`
	//
	// +private
	ReleaseDryRun bool
}

// New creates a new Ghrelease instance.
func New(
	// GitHub token with permissions to create releases
	token *dagger.Secret,
) *Ghrelease {
	return &Ghrelease{
		Token: token,
	}
}

// WithSource sets the git source directory used by Create to inspect
// tags and commit history. The directory must include .git
// (e.g. --source=. from the repository root).
func (m *Ghrelease) WithSource(
	// A directory containing a git repository (must include .git)
	//
	// +defaultPath="/"
	source *dagger.Directory,
) *Ghrelease {
	m.Source = source
	return m
}

// WithAssets sets the assets directory for release upload
func (m *Ghrelease) WithAssets(
	assets *dagger.Directory,
) *Ghrelease {
	m.Assets = assets
	return m
}

// WithRepo sets the "org/repo" repo string for release target
func (m *Ghrelease) WithRepo(
	repo string,
) *Ghrelease {
	m.Repo = repo
	return m
}

// WithFlatten enables flattening of the assets directory before upload.
// When chained, the <os>/<arch>/<filename> directory structure is collapsed
// into a flat directory with files renamed to <filename>-<os>-<arch>
// (or <filename>-<os>-<arch>.sha256 for checksum files).
func (m *Ghrelease) WithFlatten() *Ghrelease {
	m.FlattenAssets = true
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

// WithDryRun enables dry-run mode. When chained before Create, all version
// calculation and release note generation runs as normal, but the actual
// gh release create call is skipped. Useful for smoke-testing.
func (m *Ghrelease) WithDryRun() *Ghrelease {
	m.ReleaseDryRun = true
	return m
}

// Create inspects the git history in Source to determine the next semantic
// version, generates categorized release notes, and creates a new GitHub
// release. Commits since the last tag are classified as follows:
//
//   - "⚠️ breaking:" → major bump
//   - "✨ feat:" → minor bump
//   - "🔧 fix:" and everything else → patch bump
//
// The highest-priority bump wins. WithSource must be called before Create.
// Chain WithDryRun before Create to skip the actual release creation.
func (m *Ghrelease) Create(ctx context.Context) (string, error) {
	if m.Source == nil {
		return "", fmt.Errorf("no source set: call WithSource before Create")
	}

	dryRun := "false"
	if m.ReleaseDryRun {
		dryRun = "true"
	}

	out, err := dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "github-cli", "git"}).
		WithSecretVariable("GH_TOKEN", m.Token).
		WithEnvVariable("GH_REPO", m.Repo).
		WithEnvVariable("DRY_RUN", dryRun).
		WithDirectory("/src", m.Source).
		WithWorkdir("/src").
		WithNewFile("/usr/local/bin/create-release.sh", createReleaseScript, dagger.ContainerWithNewFileOpts{Permissions: 0o755}).
		WithExec([]string{"/usr/local/bin/create-release.sh"}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to create release: %w", err)
	}

	return out, nil
}

// Upload uploads all assets to a GitHub release.
// If WithFlatten was chained, the assets are flattened first.
// The tag must have been set via WithTag before calling Upload.
func (m *Ghrelease) Upload(ctx context.Context) error {
	if m.Tag == "" {
		return fmt.Errorf("no tag set: call WithTag before Upload")
	}

	dist := m.Assets

	if m.FlattenAssets {
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
