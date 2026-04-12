package git

import (
	"fmt"
	"strconv"
	"strings"
)

type ChangedFileStat struct {
	Path				string
	Additions		int
	Deletions		int
}

func GetChangedFileStats() ([]ChangedFileStat, error) {
	out, err := runGit("diff", "--numstat")
	if err != nil {
		return nil, fmt.Errorf("get unstaged diff stats: %w", err)
	}

	cachedOut, err := runGit("diff", "--cached", "--numstat")
	if err != nil {
		return nil, fmt.Errorf("get staged diff stats: %w", err)
	}

	combined := combineNumstatOutputs(out, cachedOut)
	if strings.TrimSpace(combined) == "" {
		return []ChangedFileStat{}, nil
	}

	statsMap := make(map[string]ChangedFileStat)
	lines := strings.Split(strings.TrimSpace(combined), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 3 {
			continue
		}

		additions, err := parseNumstatValue(parts[0])
		if err != nil {
			return nil, fmt.Errorf("parse additions for line %q: %w", line, err)
		}

		deletions, err := parseNumstatValue(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parse deletions for line %q: %w", line, err)
		}

		path := parts[2]

		existing := statsMap[path]
		existing.Path = path
		existing.Additions += additions
		existing.Deletions += deletions
		statsMap[path] = existing
	}

	results := make([]ChangedFileStat, 0, len(statsMap))
	for _, stat := range statsMap {
		results = append(results, stat)
	}

	return results, nil
}

func parseNumstatValue(value string) (int, error) {
	if value == "-" {
		return 0, nil
	}

	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func combineNumstatOutputs(unstaged string, staged string) string {
	unstaged = strings.TrimSpace(unstaged)
	staged = strings.TrimSpace(staged)

	switch {
	case unstaged == "" && staged == "":
		return ""
	case unstaged == "":
		return staged
	case staged == "":
		return unstaged
	default:
		return unstaged + "\n" + staged
	}
}
