package review

import (
	"strings"
	"testing"
)

func TestBuildReviewPromptIncludesFilesAndDiff(t *testing.T) {
	files := []FileChange{
		{Path: "main.go", Additions: 10, Deletions: 2},
		{Path: "internal/foo/bar.go", Additions: 1, Deletions: 0},
	}
	diff := "diff --git a/main.go b/main.go\n+added line\n-removed line"

	prompt := buildReviewPrompt(files, diff, false)

	for _, want := range []string{
		"main.go",
		"internal/foo/bar.go",
		"additions=10 deletions=2",
		diff,
		"Full git diff:",
		"STATUS: READY or NEEDS_CHANGES",
		"SUMMARY\n-------\n",
		"Critical: <count>",
	} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}

	if strings.Contains(prompt, "TRUNCATED") {
		t.Error("prompt should not mention truncation when diff is not truncated")
	}
}

func TestBuildReviewPromptTruncationNote(t *testing.T) {
	prompt := buildReviewPrompt([]FileChange{{Path: "a.go"}}, "partial diff", true)

	if !strings.Contains(prompt, "TRUNCATED") {
		t.Error("prompt should warn the model when the diff was truncated")
	}
}

func TestProviderErrorUnwrap(t *testing.T) {
	inner := &ProviderError{Err: ErrNoChanges}

	if inner.Error() != ErrNoChanges.Error() {
		t.Errorf("ProviderError.Error() = %q, want %q", inner.Error(), ErrNoChanges.Error())
	}
	if inner.Unwrap() != ErrNoChanges {
		t.Error("ProviderError.Unwrap() should return the inner error")
	}
}
