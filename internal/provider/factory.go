package provider

import (
	"fmt"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
)

func New(cfg config.Config) (Provider, error) {
	switch cfg.Provider {
	case config.ProviderOllama:
		return NewOllamaProvider(cfg.Ollama.BaseURL), nil
	case config.ProviderOpenAI:
		if cfg.OpenAI.APIKey == "" {
			return nil, fmt.Errorf("OpenAI API key is missing.\nSet it with:\n  moan config set-openai-key <key>\nor:\n  export OPENAI_API_KEY=<key>")
		}
		return NewOpenAIProvider(cfg.OpenAI.BaseURL, cfg.OpenAI.APIKey), nil
	case config.ProviderAnthropic:
		if cfg.Anthropic.APIKey == "" {
			return nil, fmt.Errorf("Anthropic API key is missing.\nSet it with:\n  moan config set-anthropic-key <key>\nor:\n  export ANTHROPIC_API_KEY=<key>")
		}
		return NewAnthropicProvider(cfg.Anthropic.BaseURL, cfg.Anthropic.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %q (supported: ollama, openai, anthropic)", cfg.Provider)
	}
}
