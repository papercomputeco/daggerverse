// Dagger module for validating GitHub pull request conventions
//
// This module provides functions to check that pull requests conform to
// project standards. It uses the GitHub CLI (gh) to inspect PR metadata
// and validates titles, labels, and other properties.

package main

import "dagger/gh-contrib/internal/dagger"

type GhContrib struct {
	// GitHub token
	//
	// +private
	Token *dagger.Secret

	// Github repo ("owner/repo")
	//
	// +private
	Repo string
}

func New(
	// GitHub token.
	token *dagger.Secret,

	// GitHub repository (e.g. "owner/repo").
	repo string,
) (*GhContrib, error) {
	return &GhContrib{
		Token: token,
		Repo:  repo,
	}, nil
}

// ghContainer returns a container with the GitHub CLI installed and authenticated.
func (m *GhContrib) ghContainer() *dagger.Container {
	return dag.Container().
		From("alpine:latest").
		WithExec([]string{"apk", "add", "--no-cache", "github-cli"}).
		WithSecretVariable("GH_TOKEN", m.Token)
}
