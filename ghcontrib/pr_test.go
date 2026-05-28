package main

import "testing"

func TestValidatePullRequestTitleAcceptsValidTitles(t *testing.T) {
	t.Parallel()

	tests := []string{
		"✨ feat: add the percolator",
		"✨ feat!: replace the percolator",
		"✨ feat(percolator): add the percolator",
		"✨ feat(percolator)!: add the percolator",
		":sparkles: feat(percolator): add the percolator",
		"🔧 fix: the build was not working on tuesdays",
		"🔧 fix(percolator): the build was not working on tuesday",
		":wrench: fix(percolator): the build was not working on tuesday",
		"🧹 chore(ci): update workflow dependencies",
		"♻️ refactor(api.v2): simplify request routing",
		"🎨 design(console-ui): tighten spacing",
		"📚 docs(readme): describe installation",
		"✏️ RFD(agent-runtime): propose sandbox lifecycle",
	}

	for _, title := range tests {
		title := title
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			if err := validatePullRequestTitle(title, 42); err != nil {
				t.Fatalf("expected title to be valid: %v", err)
			}
		})
	}
}

func TestValidatePullRequestTitleRejectsInvalidTitles(t *testing.T) {
	t.Parallel()

	tests := []string{
		"feat: add the percolator",
		"🔧 feat: add the percolator",
		"✨ fix: add the percolator",
		"🔧 fix(percolator):",
		"🔧 fix(): add the percolator",
		"🔧 fix(Percolator): add the percolator",
		"🔧 fix(percolator) add the percolator",
		"🔧 fix(percolator/subsystem): add the percolator",
		"🐛 fix: add the percolator",
		":bug: fix: add the percolator",
	}

	for _, title := range tests {
		title := title
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			if err := validatePullRequestTitle(title, 42); err == nil {
				t.Fatal("expected title to be invalid")
			}
		})
	}
}
