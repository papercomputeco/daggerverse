package main

import (
	"context"
	"dagger/go/internal/dagger"
	"errors"
	"fmt"
)

type Go struct {
	// +private
	Source *dagger.Directory
}

func New(
	// +defaultPath="/"
	source *dagger.Directory,
) *Go {
	return &Go{
		Source: source,
	}
}

func (g *Go) goContainer() *dagger.Container {
	return dag.Container().
		From("golang:1.26-bookworm").
		WithMountedCache("/go/pkg/mod", dag.CacheVolume("go-mod")).
		WithMountedCache("/root/.cache/go-build", dag.CacheVolume("go-build")).
		WithWorkdir("/src").
		WithDirectory("/src", g.Source)
}

// CheckGoModTidy runs "go mod tidy" and fails if it produces any changes to
// go.mod or go.sum, indicating that the caller forgot to tidy before committing.
//
// +check
func (g *Go) CheckGoModTidy(ctx context.Context) (string, error) {
	out, err := g.goContainer().
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
	return g.goContainer().
		WithExec([]string{"go", "vet"}).
		Stdout(ctx)
}
