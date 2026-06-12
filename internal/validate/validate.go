package validate

import (
	"context"
	"fmt"
	"time"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/git"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/provider"
)

const connectionTestTimeout = 15 * time.Second

type Item struct {
	Label string
	Value string
	OK    bool
}

type Result struct {
	Valid bool
	Items []Item
}

// Run performs the full pre-review check list: git state, config,
// provider/model selection, API key presence, and provider connectivity.
func Run(cfg config.Config) Result {
	var items []Item
	ok := func(label, value string) { items = append(items, Item{Label: label, Value: value, OK: true}) }
	fail := func(label, value string) { items = append(items, Item{Label: label, Value: value, OK: false}) }

	state, err := git.GetState()
	switch {
	case err != nil:
		fail("Git repository", fmt.Sprintf("FAIL - %v", err))
	case !state.InsideRepo:
		fail("Git repository", "FAIL - current directory is not a git repository")
	case !state.HasHead:
		fail("Git repository", "FAIL - repository has no commits yet")
	default:
		ok("Git repository", "OK")
	}

	if state.InsideRepo && state.HasHead {
		if state.StagedChanges || state.UnstagedChanges || state.UntrackedFiles {
			ok("Git changes", fmt.Sprintf("OK (%d files)", len(state.ChangedFiles)))
		} else {
			fail("Git changes", "FAIL - no changes found. Modify files or stage changes before running moan review")
		}
	}

	configPath, pathErr := config.Path()
	switch {
	case pathErr != nil:
		fail("Config", fmt.Sprintf("FAIL - %v", pathErr))
	case config.Exists():
		ok("Config", fmt.Sprintf("OK (%s)", configPath))
	default:
		ok("Config", fmt.Sprintf("OK (no file at %s, using defaults)", configPath))
	}

	providerOK := config.SupportedProvider(cfg.Provider)
	if providerOK {
		ok("Provider", cfg.Provider)
	} else {
		fail("Provider", fmt.Sprintf("FAIL - unsupported provider %q (supported: ollama, openai, anthropic)", cfg.Provider))
	}

	model := cfg.ResolvedModel()
	if model != "" {
		ok("Model", model)
	} else {
		fail("Model", "FAIL - no model configured. Set one with: moan config set-model <model>")
	}

	keyOK := true
	switch cfg.Provider {
	case config.ProviderOllama:
		ok("API key", "not required")
	case config.ProviderOpenAI:
		if cfg.OpenAI.APIKey != "" {
			ok("API key", "OK")
		} else {
			keyOK = false
			fail("API key", "FAIL - missing. Set it with: moan config set-openai-key <key> or export OPENAI_API_KEY=<key>")
		}
	case config.ProviderAnthropic:
		if cfg.Anthropic.APIKey != "" {
			ok("API key", "OK")
		} else {
			keyOK = false
			fail("API key", "FAIL - missing. Set it with: moan config set-anthropic-key <key> or export ANTHROPIC_API_KEY=<key>")
		}
	}

	if providerOK && keyOK {
		p, err := provider.New(cfg)
		if err != nil {
			fail("Connection", fmt.Sprintf("FAIL - %v", err))
		} else {
			ctx, cancel := context.WithTimeout(context.Background(), connectionTestTimeout)
			defer cancel()

			if err := p.TestConnection(ctx); err != nil {
				fail("Connection", fmt.Sprintf("FAIL - %v", err))
			} else {
				ok("Connection", "OK")
			}
		}
	} else {
		fail("Connection", "skipped (fix provider/key issues first)")
	}

	valid := true
	for _, item := range items {
		if !item.OK {
			valid = false
			break
		}
	}

	return Result{Valid: valid, Items: items}
}
