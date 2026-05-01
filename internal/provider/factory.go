package provider

import (
	"fmt"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
)

func New(cfg config.Config) (Provider, error) {
	switch cfg.Provider {
	case "ollama":
		return NewOllamaProvider(cfg.BaseURL), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}
