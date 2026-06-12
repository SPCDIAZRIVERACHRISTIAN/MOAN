package review

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/git"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/provider"
)

var (
	ErrNotRepo      = errors.New("current directory is not a git repository")
	ErrNoCommits    = errors.New("repository has no commits yet; make an initial commit before running moan review")
	ErrNoChanges    = errors.New("no Git changes found. Modify files or stage changes before running moan review")
	ErrDiffTooLarge = errors.New("diff is too large for a single review")
	ErrTooManyFiles = errors.New("too many changed files for a single review")
)

// ProviderError marks failures that came from the AI provider so the CLI
// can map them to a distinct exit code.
type ProviderError struct {
	Err error
}

func (e *ProviderError) Error() string { return e.Err.Error() }
func (e *ProviderError) Unwrap() error { return e.Err }

type Options struct {
	StagedOnly    bool
	AllowTruncate bool
	DryRun        bool
}

type FileChange struct {
	Path      string
	Additions int
	Deletions int
}

type ReviewResult struct {
	Provider      string
	Model         string
	Scope         string
	Files         []FileChange
	DiffChars     int
	PromptChars   int
	Truncated     bool
	DryRun        bool
	ReviewContent string
}

func Run(cfg config.Config, opts Options) (ReviewResult, error) {
	state, err := git.GetState()
	if err != nil {
		return ReviewResult{}, fmt.Errorf("load git state: %w", err)
	}

	if !state.InsideRepo {
		return ReviewResult{}, ErrNotRepo
	}

	if !state.HasHead {
		return ReviewResult{}, ErrNoCommits
	}

	changedFiles, err := git.GetChangedFileStats(opts.StagedOnly)
	if err != nil {
		return ReviewResult{}, fmt.Errorf("get changed file stats: %w", err)
	}

	if len(changedFiles) == 0 {
		if opts.StagedOnly {
			return ReviewResult{}, fmt.Errorf("%w (no staged changes; stage files with git add or drop --staged)", ErrNoChanges)
		}
		return ReviewResult{}, ErrNoChanges
	}

	scope := "staged + unstaged"
	if opts.StagedOnly {
		scope = "staged only"
	}

	result := ReviewResult{
		Provider: cfg.Provider,
		Model:    cfg.ResolvedModel(),
		Scope:    scope,
		DryRun:   opts.DryRun,
		Files:    make([]FileChange, 0, len(changedFiles)),
	}

	for _, file := range changedFiles {
		result.Files = append(result.Files, FileChange{
			Path:      file.Path,
			Additions: file.Additions,
			Deletions: file.Deletions,
		})
	}

	if len(result.Files) > cfg.MaxFiles {
		return ReviewResult{}, fmt.Errorf("%w: %d files changed, limit is %d.\nTry staging fewer files or increasing --max-files",
			ErrTooManyFiles, len(result.Files), cfg.MaxFiles)
	}

	diffContent, err := git.GetDiffContent(opts.StagedOnly)
	if err != nil {
		return ReviewResult{}, fmt.Errorf("get diff content: %w", err)
	}

	result.DiffChars = len(diffContent)

	if len(diffContent) > cfg.MaxDiffChars {
		if !opts.AllowTruncate {
			return ReviewResult{}, fmt.Errorf("%w: diff is %d chars, limit is %d.\nTry staging fewer files or increasing --max-diff-chars, or pass --allow-truncate",
				ErrDiffTooLarge, len(diffContent), cfg.MaxDiffChars)
		}
		diffContent = diffContent[:cfg.MaxDiffChars]
		result.Truncated = true
	}

	prompt := buildReviewPrompt(result.Files, diffContent, result.Truncated)
	result.PromptChars = len(prompt)

	if opts.DryRun {
		return result, nil
	}

	p, err := provider.New(cfg)
	if err != nil {
		return ReviewResult{}, fmt.Errorf("build provider: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.TimeoutSeconds)*time.Second)
	defer cancel()

	resp, err := p.Review(ctx, provider.ReviewRequest{
		Model:        result.Model,
		SystemPrompt: cfg.SystemPrompt,
		Messages: []provider.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return ReviewResult{}, &ProviderError{Err: err}
	}

	result.ReviewContent = resp.Content
	return result, nil
}

func buildReviewPrompt(files []FileChange, diffContent string, truncated bool) string {
	var b strings.Builder

	b.WriteString("You are MOAN, a strict code review agent.\n\n")

	b.WriteString("Your job is to review ONLY the code changes shown in the provided git diff.\n")
	b.WriteString("Be strict, but not noisy. Your value is precision. A clean diff should pass cleanly.\n\n")

	b.WriteString("CORE RULES:\n")
	b.WriteString("- Only review lines, files, functions, or behavior that are directly visible in the diff.\n")
	b.WriteString("- Do not invent missing files, hidden behavior, architecture, dependencies, or runtime context.\n")
	b.WriteString("- Do not give generic best-practice advice.\n")
	b.WriteString("- Do not complain about formatting, prompt wording, naming, structure, or verbosity unless it creates a concrete bug, security risk, broken behavior, or serious maintainability problem directly visible in the diff.\n")
	b.WriteString("- Do not classify something as a security issue unless there is a concrete security risk visible in the diff.\n")
	b.WriteString("- Prefer zero findings over weak findings.\n")
	b.WriteString("- If the diff looks safe and correct, say the review is READY with no findings.\n\n")

	b.WriteString("VALID FINDING REQUIREMENTS:\n")
	b.WriteString("A finding is valid only if it includes:\n")
	b.WriteString("1. A specific file from the diff.\n")
	b.WriteString("2. A concrete issue caused by the changed code.\n")
	b.WriteString("3. Evidence from the diff.\n")
	b.WriteString("4. A realistic impact.\n")
	b.WriteString("5. A suggested fix.\n\n")

	b.WriteString("INVALID FINDINGS:\n")
	b.WriteString("Do NOT report a finding if:\n")
	b.WriteString("- It is only a style preference.\n")
	b.WriteString("- It is only about verbosity.\n")
	b.WriteString("- It is only about naming.\n")
	b.WriteString("- It is only about possible future architecture.\n")
	b.WriteString("- It assumes code not shown in the diff.\n")
	b.WriteString("- It cannot point to changed behavior.\n")
	b.WriteString("- It cannot explain a real impact.\n")
	b.WriteString("- It is general advice that could apply to any project.\n\n")

	b.WriteString("REVIEW CATEGORIES:\n")
	b.WriteString("Use only these categories:\n")
	b.WriteString("- bug\n")
	b.WriteString("- security\n")
	b.WriteString("- reliability\n")
	b.WriteString("- performance\n")
	b.WriteString("- maintainability\n")
	b.WriteString("- test\n")
	b.WriteString("- documentation\n\n")

	b.WriteString("SEVERITY RULES:\n")
	b.WriteString("- critical: breaks core functionality, causes data loss, or introduces severe security risk.\n")
	b.WriteString("- high: likely bug, security issue, or major broken behavior.\n")
	b.WriteString("- medium: real issue with meaningful impact, but not immediately catastrophic.\n")
	b.WriteString("- low: minor but concrete issue.\n")
	b.WriteString("- info: useful note only if directly tied to the diff.\n\n")

	b.WriteString("OUTPUT FORMAT:\n\n")

	b.WriteString("STATUS: READY or NEEDS_CHANGES\n\n")

	b.WriteString("SUMMARY:\n")
	b.WriteString("Briefly summarize the diff in 1-3 sentences.\n\n")

	b.WriteString("FINDINGS:\n")
	b.WriteString("If there are no valid findings, write:\n")
	b.WriteString("No valid findings.\n\n")

	b.WriteString("If there are findings, use this exact format:\n\n")
	b.WriteString("### Finding 1\n")
	b.WriteString("File: <file path>\n")
	b.WriteString("Severity: critical | high | medium | low | info\n")
	b.WriteString("Category: bug | security | reliability | performance | maintainability | test | documentation\n")
	b.WriteString("Issue: <specific issue caused by the changed code>\n")
	b.WriteString("Evidence: <quote or describe the exact changed line or behavior from the diff>\n")
	b.WriteString("Impact: <realistic impact of the issue>\n")
	b.WriteString("Suggested fix: <concrete fix>\n\n")

	b.WriteString("FINAL DECISION:\n")
	b.WriteString("Briefly explain why the diff is READY or NEEDS_CHANGES.\n\n")

	b.WriteString("Your output MUST end with this exact block, with real counts:\n\n")
	b.WriteString("SUMMARY\n")
	b.WriteString("-------\n")
	b.WriteString("Status: READY or NEEDS WORK\n")
	b.WriteString("Critical: <count>\n")
	b.WriteString("High: <count>\n")
	b.WriteString("Medium: <count>\n")
	b.WriteString("Low: <count>\n\n")

	if truncated {
		b.WriteString("NOTE: The git diff below was TRUNCATED because it exceeded the size limit. Review what is visible and do not guess about missing content.\n\n")
	}

	b.WriteString("Changed files:\n")
	for _, f := range files {
		fmt.Fprintf(&b, "- %s | additions=%d deletions=%d\n", f.Path, f.Additions, f.Deletions)
	}

	if diffContent != "" {
		b.WriteString("\nFull git diff:\n")
		b.WriteString(diffContent)
	}

	return b.String()
}
