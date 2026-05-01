package validate

import (
	"fmt"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/git"
)

type Check struct {
	Name    string
	Passed  bool
	Message string
}

type Result struct {
	Valid  bool
	Checks []Check
}

func Run() (Result, error) {
	repoState, err := git.GetState()
	if err != nil {
		return Result{}, fmt.Errorf("load git state: %w", err)
	}

	checks := []Check{
		checkInsideRepo(repoState),
		checkHasHead(repoState),
		checkHasChanges(repoState),
	}

	if repoState.InsideRepo {
		cfgCheck := checkConfig()
		checks = append(checks, cfgCheck)
	}

	valid := true
	for _, check := range checks {
		if !check.Passed {
			valid = false
			break
		}
	}

	return Result{
		Valid:  valid,
		Checks: checks,
	}, nil
}

func checkInsideRepo(state git.State) Check {
	if state.InsideRepo {
		return Check{
			Name:    "git-repository",
			Passed:  true,
			Message: "inside a git repository",
		}
	}

	return Check{
		Name:    "git-repository",
		Passed:  false,
		Message: "current directory is not a git repository",
	}
}

func checkHasHead(state git.State) Check {
	if !state.InsideRepo {
		return Check{
			Name:    "repository-head",
			Passed:  false,
			Message: "cannot verify HEAD outside a git repository",
		}
	}

	if state.HasHead {
		return Check{
			Name:    "repository-head",
			Passed:  true,
			Message: "repository has at least one commit",
		}
	}

	return Check{
		Name:    "repository-head",
		Passed:  false,
		Message: "repository has no commits yet; HEAD is missing",
	}
}

func checkHasChanges(state git.State) Check {
	if !state.InsideRepo {
		return Check{
			Name:    "workspace-changes",
			Passed:  false,
			Message: "cannot inspect changes outside a git repository",
		}
	}

	hasChanges := state.StagedChanges || state.UnstagedChanges || state.UntrackedFiles
	if hasChanges {
		return Check{
			Name:    "workspace-changes",
			Passed:  true,
			Message: fmt.Sprintf("changes detected (%d files)", len(state.ChangedFiles)),
		}
	}

	return Check{
		Name:    "workspace-changes",
		Passed:  false,
		Message: "no staged, unstaged, or untracked changes detected",
	}
}

func checkConfig() Check {
	cfg, err := config.Load()
	if err != nil {
		return Check{
			Name:    "config",
			Passed:  false,
			Message: err.Error(),
		}
	}

	if err := config.Validate(cfg); err != nil {
		return Check{
			Name:    "config",
			Passed:  false,
			Message: err.Error(),
		}
	}

	return Check{
		Name:    "config",
		Passed:  true,
		Message: "configuration is valid",
	}
}
