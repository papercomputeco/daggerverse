package main

import (
	"fmt"
	"regexp"
	"strings"
)

type prTitleSpec struct {
	Type   string
	Tokens []string
}

type prTitlePattern struct {
	Format string
	Regexp *regexp.Regexp
}

// validPRTitleSpecs defines the allowed PR title emoji/type pairs.
var validPRTitleSpecs = []prTitleSpec{
	{Type: "feat", Tokens: []string{"✨", ":sparkles:"}},
	{Type: "fix", Tokens: []string{"🔧", ":wrench:"}},
	{Type: "chore", Tokens: []string{"🧹", ":broom:"}},
	{Type: "refactor", Tokens: []string{"♻️", ":recycle:"}},
	{Type: "design", Tokens: []string{"🎨", ":art:"}},
	{Type: "docs", Tokens: []string{"📚", ":books:"}},
	{Type: "RFD", Tokens: []string{"✏️", ":pencil2:"}},
}

var validPRTitlePatterns = buildPRTitlePatterns(validPRTitleSpecs)

func validatePullRequestTitle(title string, number int) error {
	for _, pattern := range validPRTitlePatterns {
		if pattern.Regexp.MatchString(title) {
			return nil
		}
	}

	formatList := make([]string, 0, len(validPRTitlePatterns))
	for _, pattern := range validPRTitlePatterns {
		formatList = append(formatList, "  - "+pattern.Format)
	}

	return fmt.Errorf(
		"PR #%d title %q does not match the required title format.\n\nTitle must follow: <emoji> <type>[(scope)]: <description>\nAllowed formats:\n%s",
		number,
		title,
		strings.Join(formatList, "\n"),
	)
}

func buildPRTitlePatterns(specs []prTitleSpec) []prTitlePattern {
	patterns := make([]prTitlePattern, 0, len(specs)*2)
	for _, spec := range specs {
		for _, token := range spec.Tokens {
			patterns = append(patterns, prTitlePattern{
				Format: fmt.Sprintf("%s %s[(scope)]: description", token, spec.Type),
				Regexp: regexp.MustCompile(fmt.Sprintf(
					`^%s %s(?:\([a-z0-9][a-z0-9._-]*\))?!?: .+$`,
					regexp.QuoteMeta(token),
					regexp.QuoteMeta(spec.Type),
				)),
			})
		}
	}
	return patterns
}
