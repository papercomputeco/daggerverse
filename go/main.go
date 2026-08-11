package main

import (
	"context"
	"dagger/go/internal/dagger"
	"errors"
	"fmt"
	"strings"
)

type Go struct {
	// +private
	Source *dagger.Directory

	// EnvVars is an optional list of environment variables to set in the
	// Go container. Each entry must be in "KEY=VALUE" format.
	// This is useful for build-time variables like GOEXPERIMENT.
	//
	// +private
	EnvVars []string
}

func New(
	// +defaultPath="/"
	source *dagger.Directory,

	// Optional environment variables to set in the Go container.
	// Each entry must be in "KEY=VALUE" format (e.g. "GOEXPERIMENT=jsonv2").
	// +optional
	envVars []string,
) *Go {
	return &Go{
		Source:  source,
		EnvVars: envVars,
	}
}

func (g *Go) goContainer() (*dagger.Container, error) {
	ctr := dag.Container().
		From("golang:1.26-bookworm").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir("/src").
		WithDirectory("/src", g.Source)

	// Apply caller-provided environment variables.
	for _, env := range g.EnvVars {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return nil, fmt.Errorf("invalid env var %q: must be in KEY=VALUE format", env)
		}
		ctr = ctr.WithEnvVariable(parts[0], parts[1])
	}

	return ctr, nil
}

// CheckGoModTidy runs "go mod tidy" and fails if it produces any changes to
// go.mod or go.sum, indicating that the caller forgot to tidy before committing.
//
// +check
func (g *Go) CheckGoModTidy(ctx context.Context) (string, error) {
	ctr, err := g.goContainer()
	if err != nil {
		return "", fmt.Errorf("could not create go container: %w", err)
	}

	out, err := ctr.
		WithExec([]string{"cp", "go.mod", "go.mod.HEAD"}).
		WithExec([]string{"cp", "go.sum", "go.sum.HEAD"}).
		WithExec([]string{"go", "mod", "tidy"}).
		WithExec([]string{
			"sh", "-c",
			"diff -u go.mod.HEAD go.mod && diff -u go.sum.HEAD go.sum",
		}).
		Stdout(ctx)

	var e *dagger.ExecError
	if errors.As(err, &e) {
		return "", fmt.Errorf(
			"go.mod or go.sum are not tidy: run 'go mod tidy' and commit the changes\n\n%s",
			e.Stdout,
		)
	} else if err != nil {
		return "", fmt.Errorf("unexpected error: %w", err)
	}

	return fmt.Sprintf("go.mod and go.sum are tidy: %s", out), nil
}

// CheckGoVet runs "go vet" against the Source directory and the root Go mod
//
// +check
func (g *Go) CheckGoVet(ctx context.Context) (string, error) {
	ctr, err := g.goContainer()
	if err != nil {
		return "", fmt.Errorf("could not create go container: %w", err)
	}

	return ctr.
		WithExec([]string{"go", "vet", "./..."}).
		Stdout(ctx)
}
