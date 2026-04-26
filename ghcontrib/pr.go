package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
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

// Linear's magic words compiled from:
// https://linear.app/docs/github#link-through-pull-requests
var linearMagicWordPattern = regexp.MustCompile(`(?i)\b(?:close|closes|closed|closing|closing fix|fix|fixes|fixed|fixing|resolve|resolves|resolved|resolving|complete|completes|completed|completing|implements|implemented|implementing|ref|refs|references|part of|related to|contributes to|toward|towards)\s+(?-i:(?:PCC|DES|REL|CTO)-[0-9]+)\b`)

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

	if err := validatePullRequestTitle(pr.Title, number); err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ PR #%d title is valid: %q", number, pr.Title), nil
}

// CheckPullRequestLinearMagicWord validates that a pull request title or body
// references a Linear issue using a GitHub magic word, e.g. "fixes PCC-123".
func (m *Ghcontrib) CheckPullRequestLinearMagicWord(
	ctx context.Context,

	// The pull request number to check
	number int,
) (string, error) {
	prJSON, err := m.ghContainer().
		WithExec([]string{
			"gh", "pr", "view",
			fmt.Sprintf("%d", number),
			"--repo", m.Repo,
			"--json", "title,body",
		}).
		Stdout(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch PR #%d from %s: %w", number, m.Repo, err)
	}

	var pr struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(prJSON)), &pr); err != nil {
		return "", fmt.Errorf("failed to parse PR JSON response: %w", err)
	}

	if err := validatePullRequestLinearMagicWord(pr.Title, pr.Body, number); err != nil {
		return "", err
	}

	return fmt.Sprintf("✅ PR #%d has a valid Linear magic word", number), nil
}

func validatePullRequestTitle(title string, number int) error {
	for _, prefix := range validPRPrefixes {
		if strings.HasPrefix(title, prefix) {
			return nil
		}
	}

	prefixList := make([]string, len(validPRPrefixes))
	for i, p := range validPRPrefixes {
		prefixList[i] = fmt.Sprintf("  - %q", p)
	}

	return fmt.Errorf(
		"PR #%d title %q does not match any required prefix.\n\nTitle must start with one of:\n%s",
		number,
		title,
		strings.Join(prefixList, "\n"),
	)
}

func validatePullRequestLinearMagicWord(title, body string, number int) error {
	if linearMagicWordPattern.MatchString(title) || linearMagicWordPattern.MatchString(body) {
		return nil
	}

	return fmt.Errorf(
		"PR #%d does not reference a Linear issue with a required magic word.\n\nExpected format: <magic word> <Linear team>-123\nAllowed Linear teams: PCC, DES, REL\nExamples: fixes PCC-123, related to DES-456",
		number,
	)
}
