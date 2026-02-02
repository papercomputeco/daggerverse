package main

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"dagger/golangcilint/internal/dagger"
)

const (
	golangciLintImage string = "golangci/golangci-lint:v2.8.0"
)

//go:embed .golangci.yml
var defaultConfig string

type Golangcilint struct {
	// Source is the Go source directory to lint.
	//
	// +private
	Source *dagger.Directory

	// Config is an optional golangci-lint configuration file.
	// When provided it overrides the opinionated default config that ships
	// with this module.
	//
	// +private
	Config *dagger.File

	// EnvVars is an optional list of environment variables to set in the
	// lint container. Each entry must be in "KEY=VALUE" format.
	// This is useful for build-time variables like GOEXPERIMENT.
	//
	// +private
	EnvVars []string

	// BaseCtr is an optional base container with golangci-lint already installed.
	// When provided it replaces the default golangci-lint image, allowing
	// callers to supply extra system libraries or tooling (e.g. sqlite-dev).
	// The container must have golangci-lint on PATH.
	//
	// +private
	BaseCtr *dagger.Container
}

// New creates a new Golangcilint module instance.
func New(
	// The Go source directory to lint.
	source *dagger.Directory,

	// An optional golangci-lint configuration file (.golangci.yml) that
	// replaces the built-in opinionated defaults.
	// +optional
	config *dagger.File,

	// Optional environment variables to set in the lint container.
	// Each entry must be in "KEY=VALUE" format (e.g. "GOEXPERIMENT=rangefunc").
	// +optional
	envVars []string,

	// Optional base container with golangci-lint already installed.
	// When provided it replaces the default golangci-lint image, allowing
	// callers to supply extra system libraries or tooling (e.g. sqlite-dev).
	// The container must have golangci-lint on PATH.
	// +optional
	baseCtr *dagger.Container,
) *Golangcilint {
	return &Golangcilint{
		Source:  source,
		Config:  config,
		EnvVars: envVars,
		BaseCtr: baseCtr,
	}
}

// Lint runs golangci-lint on the source directory with --fix, applying
// auto-fixes where possible, and returns the directory with fixes applied.
func (m *Golangcilint) Lint() (*dagger.Directory, error) {
	ctr, err := m.lintContainer()
	if err != nil {
		return nil, fmt.Errorf("could not create lint container: %w", err)
	}

	return ctr.
		WithExec(m.buildArgs("--fix")).
		Directory("/src"), nil
}

// Check runs golangci-lint on the source directory without applying fixes.
// It returns the linter output as a string. If there are lint violations the
// Dagger pipeline will fail, making this suitable for CI checks.
func (m *Golangcilint) Check(ctx context.Context) (string, error) {
	ctr, err := m.lintContainer()
	if err != nil {
		return "", fmt.Errorf("could not create lint container: %w", err)
	}

	return ctr.
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
func (m *Golangcilint) lintContainer() (*dagger.Container, error) {
	var ctr *dagger.Container
	if m.BaseCtr != nil {
		ctr = m.BaseCtr
	} else {
		ctr = dag.Container().From(golangciLintImage)
	}

	ctr = ctr.
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

	// Apply caller-provided environment variables.
	for _, env := range m.EnvVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid env var %q: must be in KEY=VALUE format", env)
		}
		ctr = ctr.WithEnvVariable(parts[0], parts[1])
	}

	return ctr, nil
}
