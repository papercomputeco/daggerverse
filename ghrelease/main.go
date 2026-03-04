// GitHub release management.
//
// Create looks at git history to determine the next semver tag, generates
// categorized release notes, and creates a GitHub release via the gh CLI.

package main

import (
	"context"
	_ "embed"
	"fmt"

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

	// When true, run all logic but skip the actual gh release create
	//
	// +private
	DryRun bool
}

// New creates a new Ghrelease instance.
func New(
	// GitHub token with permissions to create releases
	token *dagger.Secret,

	// GitHub repository in owner/repo format (e.g., "papercomputeco/myproject")
	repo string,
) *Ghrelease {
	return &Ghrelease{
		Token: token,
		Repo:  repo,
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

// WithDryRun enables dry-run mode. When chained before Create, all version
// calculation and release note generation runs as normal, but the actual
// gh release create call is skipped. Useful for smoke-testing.
func (m *Ghrelease) WithDryRun() *Ghrelease {
	m.DryRun = true
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
	if m.DryRun {
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
