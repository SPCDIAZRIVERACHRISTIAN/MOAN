package review

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/git"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/provider"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/validate"
)

type FileChange struct {
	Path      string
	Additions int
	Deletions int
}

type ReviewResult struct {
	Ready         bool
	Provider      string
	Model         string
	Files         []FileChange
	ReviewContent string
}

func Run() (ReviewResult, error) {
	validationResult, err := validate.Run()
	if err != nil {
		return ReviewResult{}, fmt.Errorf("run validation: %w", err)
	}

	if !validationResult.Valid {
		return ReviewResult{
			Ready: false,
			Files: []FileChange{},
		}, nil
	}

	cfg, err := config.Load()
	if err != nil {
		return ReviewResult{}, fmt.Errorf("load config: %w", err)
	}

	changedFiles, err := git.GetChangedFileStats()
	if err != nil {
		return ReviewResult{}, fmt.Errorf("get changed file stats: %w", err)
	}

	result := ReviewResult{
		Ready:    len(changedFiles) > 0,
		Provider: cfg.Provider,
		Model:    cfg.Model,
		Files:    make([]FileChange, 0, len(changedFiles)),
	}

	for _, file := range changedFiles {
		result.Files = append(result.Files, FileChange{
			Path:      file.Path,
			Additions: file.Additions,
			Deletions: file.Deletions,
		})
	}

	if !result.Ready {
		return result, nil
	}

	p, err := provider.New(cfg)
	if err != nil {
		return ReviewResult{}, fmt.Errorf("build provider: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	diffContent, err := git.GetDiffContent()
	if err != nil {
		return ReviewResult{}, fmt.Errorf("get diff content: %w", err)
	}

	prompt := buildReviewPrompt(result.Files, diffContent)

	resp, err := p.Review(ctx, provider.ReviewRequest{
		Model:        cfg.Model,
		SystemPrompt: cfg.SystemPrompt,
		Messages: []provider.Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	})
	if err != nil {
		return ReviewResult{}, fmt.Errorf("run model review: %w", err)
	}

	result.ReviewContent = resp.Content
	return result, nil
}

func buildReviewPrompt(files []FileChange, diffContent string) string {
	var b strings.Builder

	b.WriteString("You are reviewing a git diff.\n\n")
	b.WriteString("Review rules:\n")
	b.WriteString("- Only review changes that are directly shown in the provided diff.\n")
	b.WriteString("- Do not invent files, functions, risks, or behavior that are not visible in the diff.\n")
	b.WriteString("- Do not give generic best-practice advice unless it is tied to a specific changed line or behavior.\n")
	b.WriteString("- Do not classify something as a security issue unless there is a concrete security risk in the diff.\n")
	b.WriteString("- Prefer fewer high-quality findings over many weak findings.\n")
	b.WriteString("- If the diff does not contain meaningful issues, say: \"No major issues found.\"\n\n")
	b.WriteString("For each finding, use this exact format:\n\n")
	b.WriteString("### Finding <number>\n\n")
	b.WriteString("File: <file path>\n")
	b.WriteString("Severity: critical | high | medium | low\n")
	b.WriteString("Category: bug | security | architecture | maintainability | test\n")
	b.WriteString("Issue: <specific issue found in the diff>\n")
	b.WriteString("Evidence: <quote or describe the exact changed line/behavior that supports the finding>\n")
	b.WriteString("Why it matters: <practical impact>\n")
	b.WriteString("Suggested fix: <concrete fix>\n\n")
	b.WriteString("After the findings, include:\n\n")
	b.WriteString("### Summary\n")
	b.WriteString("- <short summary of the most important risk>\n")
	b.WriteString("- <recommended next action>\n\n")
	b.WriteString("Do not include sections for files that have no meaningful issues.\n\n")

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
