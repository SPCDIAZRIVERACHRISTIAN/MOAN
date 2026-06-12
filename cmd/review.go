/*
Copyright © 2026 11b_shrink christianda3@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/review"
	"github.com/spf13/cobra"
)

var (
	reviewProvider      string
	reviewModel         string
	reviewTimeout       int
	reviewStaged        bool
	reviewOutput        string
	reviewMaxDiffChars  int
	reviewMaxFiles      int
	reviewAllowTruncate bool
	reviewDryRun        bool
	reviewDebug         bool
	reviewSave          bool
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Review local Git changes with the configured AI provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		if reviewOutput != "text" && reviewOutput != "markdown" {
			return exitWithCode(exitRuntime, fmt.Errorf("invalid --output %q (supported: text, markdown)", reviewOutput))
		}

		cfg, err := config.Load()
		if err != nil {
			return exitWithCode(exitRuntime, err)
		}

		// Flag overrides apply to this run only; the config file is never
		// rewritten here.
		if reviewProvider != "" {
			cfg.Provider = reviewProvider
		}
		if reviewModel != "" {
			cfg.Model = reviewModel
		}
		if cmd.Flags().Changed("timeout") {
			cfg.TimeoutSeconds = reviewTimeout
		}
		if cmd.Flags().Changed("max-diff-chars") {
			cfg.MaxDiffChars = reviewMaxDiffChars
		}
		if cmd.Flags().Changed("max-files") {
			cfg.MaxFiles = reviewMaxFiles
		}

		if reviewDebug {
			printReviewDebug(cfg)
		}

		opts := review.Options{
			StagedOnly:    reviewStaged,
			AllowTruncate: reviewAllowTruncate,
			DryRun:        reviewDryRun,
		}

		var stopLoader func()
		if !reviewDryRun {
			stopLoader = startReviewLoader()
		}

		result, err := review.Run(cfg, opts)

		if stopLoader != nil {
			stopLoader()
		}

		if err != nil {
			return exitWithCode(reviewExitCode(err), err)
		}

		var output string
		if result.DryRun {
			output = renderDryRun(result)
		} else if reviewOutput == "markdown" {
			output = renderMarkdown(result)
		} else {
			output = renderText(result)
		}

		fmt.Println(output)

		if reviewSave && !result.DryRun {
			path, err := saveReview(result)
			if err != nil {
				return exitWithCode(exitRuntime, fmt.Errorf("save review: %w", err))
			}
			fmt.Printf("Review saved to %s\n", path)
		}

		return nil
	},
}

func reviewExitCode(err error) int {
	var pErr *review.ProviderError

	switch {
	case errors.Is(err, review.ErrNoChanges):
		return exitNoChanges
	case errors.Is(err, review.ErrNotRepo), errors.Is(err, review.ErrNoCommits):
		return exitValidation
	case errors.As(err, &pErr):
		return exitProvider
	default:
		return exitRuntime
	}
}

func printReviewDebug(cfg config.Config) {
	configPath, _ := config.Path()

	var baseURL string
	switch cfg.Provider {
	case config.ProviderOllama:
		baseURL = cfg.Ollama.BaseURL
	case config.ProviderOpenAI:
		baseURL = cfg.OpenAI.BaseURL
	case config.ProviderAnthropic:
		baseURL = cfg.Anthropic.BaseURL
	}

	fmt.Println("DEBUG")
	fmt.Println("-----")
	fmt.Printf("Config path: %s (exists: %v)\n", configPath, config.Exists())
	fmt.Printf("Provider: %s\n", cfg.Provider)
	fmt.Printf("Model: %s\n", cfg.ResolvedModel())
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Timeout: %ds\n", cfg.TimeoutSeconds)
	fmt.Printf("Max diff chars: %d\n", cfg.MaxDiffChars)
	fmt.Printf("Max files: %d\n", cfg.MaxFiles)
	fmt.Printf("OpenAI key: %s\n", config.MaskKey(cfg.OpenAI.APIKey))
	fmt.Printf("Anthropic key: %s\n", config.MaskKey(cfg.Anthropic.APIKey))
	fmt.Println()
}

func renderDryRun(result review.ReviewResult) string {
	var b strings.Builder

	b.WriteString("DRY RUN (no provider call made)\n")
	b.WriteString("-------------------------------\n")
	fmt.Fprintf(&b, "Provider: %s\n", result.Provider)
	fmt.Fprintf(&b, "Model: %s\n", result.Model)
	fmt.Fprintf(&b, "Review scope: %s\n", result.Scope)
	fmt.Fprintf(&b, "Files included: %d\n", len(result.Files))

	for _, file := range result.Files {
		fmt.Fprintf(&b, "- %s | +%d -%d\n", file.Path, file.Additions, file.Deletions)
	}

	fmt.Fprintf(&b, "Diff size: %d chars\n", result.DiffChars)
	fmt.Fprintf(&b, "Estimated request size: %d chars\n", result.PromptChars)

	if result.Truncated {
		b.WriteString("Diff would be TRUNCATED to fit the size limit.\n")
	}

	return b.String()
}

func renderText(result review.ReviewResult) string {
	var b strings.Builder

	b.WriteString("STATUS: READY\n")
	fmt.Fprintf(&b, "Provider: %s\n", result.Provider)
	fmt.Fprintf(&b, "Model: %s\n", result.Model)
	fmt.Fprintf(&b, "Review scope: %s\n", result.Scope)
	fmt.Fprintf(&b, "Files changed: %d\n\n", len(result.Files))

	for _, file := range result.Files {
		fmt.Fprintf(&b, "- %s | +%d -%d\n", file.Path, file.Additions, file.Deletions)
	}

	if result.Truncated {
		b.WriteString("\nNOTE: diff was truncated to fit the size limit before review.\n")
	}

	b.WriteString("\nAI REVIEW\n")
	b.WriteString("-----------------\n")

	if result.ReviewContent != "" {
		b.WriteString(result.ReviewContent)
	} else {
		b.WriteString("No model response.")
	}

	return b.String()
}

func renderMarkdown(result review.ReviewResult) string {
	var b strings.Builder

	b.WriteString("# MOAN Review\n\n")
	fmt.Fprintf(&b, "- **Provider:** %s\n", result.Provider)
	fmt.Fprintf(&b, "- **Model:** %s\n", result.Model)
	fmt.Fprintf(&b, "- **Review scope:** %s\n", result.Scope)
	fmt.Fprintf(&b, "- **Date:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	b.WriteString("## Changed files\n\n")
	for _, file := range result.Files {
		fmt.Fprintf(&b, "- `%s` (+%d/-%d)\n", file.Path, file.Additions, file.Deletions)
	}

	if result.Truncated {
		b.WriteString("\n> Note: diff was truncated to fit the size limit before review.\n")
	}

	b.WriteString("\n## AI Review\n\n")

	if result.ReviewContent != "" {
		b.WriteString(result.ReviewContent)
	} else {
		b.WriteString("No model response.")
	}

	b.WriteString("\n")
	return b.String()
}

func saveReview(result review.ReviewResult) (string, error) {
	dir := filepath.Join(".moan", "reviews")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	path := filepath.Join(dir, time.Now().Format("2006-01-02-150405")+".md")
	if err := os.WriteFile(path, []byte(renderMarkdown(result)), 0o644); err != nil {
		return "", err
	}

	return path, nil
}

func init() {
	rootCmd.AddCommand(reviewCmd)

	reviewCmd.Flags().StringVar(&reviewProvider, "provider", "", "override provider for this run (ollama, openai, anthropic)")
	reviewCmd.Flags().StringVar(&reviewModel, "model", "", "override model for this run")
	reviewCmd.Flags().IntVar(&reviewTimeout, "timeout", config.DefaultTimeoutSeconds, "request timeout in seconds")
	reviewCmd.Flags().BoolVar(&reviewStaged, "staged", false, "review staged changes only")
	reviewCmd.Flags().StringVar(&reviewOutput, "output", "text", "output format (text, markdown)")
	reviewCmd.Flags().IntVar(&reviewMaxDiffChars, "max-diff-chars", config.DefaultMaxDiffChars, "maximum diff size in characters")
	reviewCmd.Flags().IntVar(&reviewMaxFiles, "max-files", config.DefaultMaxFiles, "maximum number of changed files")
	reviewCmd.Flags().BoolVar(&reviewAllowTruncate, "allow-truncate", false, "truncate oversized diffs instead of failing")
	reviewCmd.Flags().BoolVar(&reviewDryRun, "dry-run", false, "build the review request without calling the provider")
	reviewCmd.Flags().BoolVar(&reviewDebug, "debug", false, "print internal debug info (never prints full API keys)")
	reviewCmd.Flags().BoolVar(&reviewSave, "save", false, "save the review to .moan/reviews/")
}

func startReviewLoader() func() {
	done := make(chan struct{})
	cleared := make(chan struct{})

	go func() {
		defer close(cleared)
		frames := []string{"[m   ]", "[mo  ]", "[moa ]", "[moan]", "[MOAN]"}
		i := 0

		for {
			select {
			case <-done:
				fmt.Print("\r                                             \r")
				return
			default:
				fmt.Printf("\rMOAN> %s chewing through your diff...", frames[i%len(frames)])
				i++
				time.Sleep(180 * time.Millisecond)
			}
		}
	}()

	return func() {
		close(done)
		<-cleared
	}
}
