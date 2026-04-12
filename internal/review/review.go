package review

import (
	"fmt"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/git"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/validate"
)

type FileChange struct {
	Path				string
	Additions		int
	Deletions		int
}

type ReviewResult struct {
	Ready	bool
	Files	[]FileChange
}

func Run() (ReviewResult, error) {
	validationResult, error := validate.Run()
	if err != nil {
		return ReviewResult{}, fmt.Errorf("run validation: %w", err)
	}

	if !validationResult.Valid {
		return ReviewResult{
			Ready: false,
			Files: []FileChange{},
		}, nil
	}

	changedFiles,err := git.GetChangedFileStats()
	if err != nil {
		return ReviewResult{}, ftm.Error("get changed file stats: %w", err)
	}

	result := ReviewResult{
		Ready: len(changedFiles) > 0,
		Files: make([]FileChange, 0, len(changedFiles)),
	}

	for _, file := range changedFiles {
		result.Files = append(result.Files, FileChange{
			Path:				file.Path,
			Additions:	file.Additions,
			Deletions:	file.Deletions,
		})
	}

	return result, nil
}
