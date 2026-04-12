package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

type State struct {
	InsideRepo      bool
	HasHead         bool
	StagedChanges   bool
	UnstagedChanges bool
	UntrackedFiles  bool
	ChangedFiles    []string
}

func GetState() (State, error) {
	state := State{}

	insideRepo, err := isInsideRepo()
	if err != nil {
		return state, fmt.Errorf("check repo state: %w", err)
	}
	state.InsideRepo = insideRepo

	if !state.InsideRepo {
		return state, nil
	}

	hasHead, err := hasHead()
	if err != nil {
		return state, fmt.Errorf("check HEAD state: %w", err)
	}
	state.HasHead = hasHead

	staged, err := hasStagedChanges()
	if err != nil {
		return state, fmt.Errorf("check staged changes: %w", err)
	}
	state.StagedChanges = staged

	unstaged, err := hasUnstagedChanges()
	if err != nil {
		return state, fmt.Errorf("check unstaged changes: %w", err)
	}
	state.UnstagedChanges = unstaged

	untracked, err := hasUntrackedFiles()
	if err != nil {
		return state, fmt.Errorf("check untracked files: %w", err)
	}
	state.UntrackedFiles = untracked

	files, err := changedFiles()
	if err != nil {
		return state, fmt.Errorf("get changed files: %w", err)
	}
	state.ChangedFiles = files

	return state, nil
}

func isInsideRepo() (bool, error) {
	out, err := runGit("rev-parse", "--is-inside-work-tree")
	if err != nil {
		// If git says we're not in a repo, treat as false, not hard error.
		return false, nil
	}

	return strings.TrimSpace(out) == "true", nil
}

func hasHead() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	err := cmd.Run()
	if err != nil {
		return false, nil
	}
	return true, nil
}

func hasStagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	err := cmd.Run()

	if err == nil {
		return false, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return true, nil
	}

	return false, err
}

func hasUnstagedChanges() (bool, error) {
	cmd := exec.Command("git", "diff", "--quiet")
	err := cmd.Run()

	if err == nil {
		return false, nil
	}

	if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
		return true, nil
	}

	return false, err
}

func hasUntrackedFiles() (bool, error) {
	out, err := runGit("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(out) != "", nil
}

func changedFiles() ([]string, error) {
	out, err := runGit("status", "--short")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(out), "\n")
	files := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// git status --short format starts with 2 status chars, then filename
		if len(line) > 3 {
			files = append(files, strings.TrimSpace(line[3:]))
		}
	}

	return files, nil
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}

	return stdout.String(), nil
}
