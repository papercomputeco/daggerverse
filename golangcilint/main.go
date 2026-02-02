package main

import (
	"context"
	_ "embed"

	"dagger/golangcilint/internal/dagger"
)

const (
	golangciLintImage string = "golangci/golangci-lint:v2.8.0"
)

//go:embed .golangci.yml
var defaultConfig string

type Golangcilint struct {
	// Source is the Go source directory to lint.
	Source *dagger.Directory

	// Config is an optional golangci-lint configuration file.
	// When provided it overrides the opinionated default config that ships
	// with this module.
	// +optional
	Config *dagger.File
}

// New creates a new Golangcilint module instance.
func New(
	// The Go source directory to lint.
	source *dagger.Directory,

	// An optional golangci-lint configuration file (.golangci.yml) that
	// replaces the built-in opinionated defaults.
	// +optional
	config *dagger.File,
) *Golangcilint {
	return &Golangcilint{
		Source: source,
		Config: config,
	}
}

// Lint runs golangci-lint on the source directory with --fix, applying
// auto-fixes where possible, and returns the directory with fixes applied.
func (m *Golangcilint) Lint() *dagger.Directory {
	return m.lintContainer().
		WithExec(m.buildArgs("--fix")).
		Directory("/src")
}

// Check runs golangci-lint on the source directory without applying fixes.
// It returns the linter output as a string. If there are lint violations the
// Dagger pipeline will fail, making this suitable for CI checks.
func (m *Golangcilint) Check(ctx context.Context) (string, error) {
	return m.lintContainer().
		WithExec(m.buildArgs()).
		Stdout(ctx)
}

// buildArgs constructs the golangci-lint command arguments.
// Any extra flags passed in are appended after the base command.
func (m *Golangcilint) buildArgs(extra ...string) []string {
	args := []string{"golangci-lint", "run", "--config", "/src/.golangci.yml"}
	args = append(args, extra...)
	args = append(args, "./...")
	return args
}

// lintContainer returns a container configured for running golangci-lint
// with Go module and build caches, and the config file mounted.
func (m *Golangcilint) lintContainer() *dagger.Container {
	ctr := dag.Container().
		From(golangciLintImage).
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithMountedCache("/root/.cache/golangci-lint", dag.CacheVolume("golangci-lint")).
		WithWorkdir("/src").
		WithDirectory("/src", m.Source)

	// Mount either the user-supplied config or the built-in default.
	if m.Config != nil {
		ctr = ctr.WithFile("/src/.golangci.yml", m.Config)
	} else {
		ctr = ctr.WithNewFile("/src/.golangci.yml", defaultConfig)
	}

	return ctr
}
