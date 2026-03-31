package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// validPRPrefixes defines the set of allowed PR title prefixes.
var validPRPrefixes = []string{
	// Literal emojis
	"✨ feat: ",
	"🔧 fix: ",
	"🧹 chore: ",
	"♻️ refactor: ",

	// Colon based emoji derivatives
	":sparkles: feat: ",
	":wrench: fix: ",
	":broom: chore: ",
	":recycle: refactor: ",
}

// CheckPullRequest validates that a pull request conforms to project standards.
// It checks that the PR title starts with one of the required prefixes.
//
// This is intended to be called from a GitHub Actions workflow where the
// GitHub token and PR metadata are available.
func (m *Ghcontrib) CheckPullRequest(
	ctx context.Context,

	// The pull request number to check
	number int,
) (string, error) {
	// Fetch the PR title using the gh CLI
	prJSON, err := m.ghContainer().
		WithExec([]string{
			"gh", "pr", "view",
			fmt.Sprintf("%d", number),
			"--repo", m.Repo,
			"--json", "title",
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch PR #%d from %s: %w", number, m.Repo, err)
	}

	// Parse the JSON response to extract the title
	var pr struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(prJSON)), &pr); err != nil {
		return "", fmt.Errorf("failed to parse PR JSON response: %w", err)
	}

	// Validate the PR title against the allowed prefixes
	for _, prefix := range validPRPrefixes {
		if strings.HasPrefix(pr.Title, prefix) {
			return fmt.Sprintf("✅ PR #%d title is valid: %q", number, pr.Title), nil
		}
	}

	// Build the error message with the list of valid prefixes
	prefixList := make([]string, len(validPRPrefixes))
	for i, p := range validPRPrefixes {
		prefixList[i] = fmt.Sprintf("  - %q", p)
	}

	return "", fmt.Errorf(
		"PR #%d title %q does not match any required prefix.\n\nTitle must start with one of:\n%s",
		number,
		pr.Title,
		strings.Join(prefixList, "\n"),
	)
}
