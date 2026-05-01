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

	b.WriteString("Review this git diff.\n")
	b.WriteString("Focus on bugs, risky changes, architecture concerns, maintainability, and security.\n")
	b.WriteString("Only review the changes shown in the diff. Do not invent unrelated files or issues.\n")
	b.WriteString("Be concise and structured.\n\n")

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
