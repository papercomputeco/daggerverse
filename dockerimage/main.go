// Build and publish multi-platform images from Dockerfiles.
package main

import (
	"context"
	"fmt"

	"dagger/dockerimage/internal/dagger"
)

var defaultPlatforms = []dagger.Platform{
	"linux/amd64",
	"linux/arm64",
}

type buildArg struct {
	Name  string
	Value string
}

// Dockerimage builds and publishes a Dockerfile from a source directory.
type Dockerimage struct {
	// +private
	Source *dagger.Directory

	// +private
	Dockerfile string

	// +private
	BuildArgs []buildArg

	// +private
	Platforms []dagger.Platform
}

// New configures a Dockerfile image workflow.
func New(
	// Build context.
	// +defaultPath="/"
	source *dagger.Directory,

	// Dockerfile path relative to the build context.
	// +optional
	// +default="Dockerfile"
	dockerfile string,

	// Platforms included when publishing.
	// +optional
	platforms []dagger.Platform,
) *Dockerimage {
	if len(platforms) == 0 {
		platforms = defaultPlatforms
	}

	return &Dockerimage{
		Source:     source,
		Dockerfile: dockerfile,
		Platforms:  platforms,
	}
}

// WithBuildArg adds a Docker build argument.
func (m *Dockerimage) WithBuildArg(name, value string) *Dockerimage {
	m.BuildArgs = append(m.BuildArgs, buildArg{Name: name, Value: value})
	return m
}

// Build builds the image for the local platform, or for platform when provided.
func (m *Dockerimage) Build(
	// Target platform, for example linux/arm64.
	// +optional
	platform dagger.Platform,
) *dagger.Container {
	buildArgs := make([]dagger.BuildArg, 0, len(m.BuildArgs))
	for _, arg := range m.BuildArgs {
		buildArgs = append(buildArgs, dagger.BuildArg{Name: arg.Name, Value: arg.Value})
	}

	return m.Source.DockerBuild(dagger.DirectoryDockerBuildOpts{
		Dockerfile: m.Dockerfile,
		BuildArgs:  buildArgs,
		Platform:   platform,
	})
}

// Publish builds all configured platform variants and publishes each tag.
func (m *Dockerimage) Publish(
	ctx context.Context,

	// Complete untagged repository, for example registry.example.com/team/app.
	repository string,

	// Tags to publish, for example ["v1.0.0", "latest"].
	tags []string,
) ([]string, error) {
	if repository == "" {
		return nil, fmt.Errorf("repository is required")
	}
	if len(tags) == 0 {
		return nil, fmt.Errorf("at least one tag is required")
	}

	variants := make([]*dagger.Container, 0, len(m.Platforms))
	for _, platform := range m.Platforms {
		variants = append(variants, m.Build(platform))
	}

	published := make([]string, 0, len(tags))
	for _, tag := range tags {
		if tag == "" {
			return published, fmt.Errorf("tag must not be empty")
		}

		ref := fmt.Sprintf("%s:%s", repository, tag)
		address, err := dag.Container().Publish(ctx, ref, dagger.ContainerPublishOpts{
			PlatformVariants: variants,
		})
		if err != nil {
			return published, fmt.Errorf("failed to publish %s: %w", ref, err)
		}
		published = append(published, address)
	}

	return published, nil
}
