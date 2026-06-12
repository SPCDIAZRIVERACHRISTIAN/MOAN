package provider

import (
	"strings"
	"testing"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
)

func baseConfig() config.Config {
	return config.Config{
		Provider: config.ProviderOllama,
		Ollama:   config.OllamaConfig{BaseURL: "http://localhost:11434"},
		OpenAI:   config.OpenAIConfig{BaseURL: "https://api.openai.com/v1"},
		Anthropic: config.AnthropicConfig{
			BaseURL: "https://api.anthropic.com",
		},
	}
}

func TestNewOllama(t *testing.T) {
	cfg := baseConfig()

	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, isOllama := p.(*OllamaProvider); !isOllama {
		t.Errorf("New() returned %T, want *OllamaProvider", p)
	}
}

func TestNewOpenAI(t *testing.T) {
	cfg := baseConfig()
	cfg.Provider = config.ProviderOpenAI

	if _, err := New(cfg); err == nil {
		t.Fatal("New() with no OpenAI key should fail")
	} else if !strings.Contains(err.Error(), "set-openai-key") {
		t.Errorf("missing-key error should mention how to fix it, got: %v", err)
	}

	cfg.OpenAI.APIKey = "sk-test"
	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, isOpenAI := p.(*OpenAIProvider); !isOpenAI {
		t.Errorf("New() returned %T, want *OpenAIProvider", p)
	}
}

func TestNewAnthropic(t *testing.T) {
	cfg := baseConfig()
	cfg.Provider = config.ProviderAnthropic

	if _, err := New(cfg); err == nil {
		t.Fatal("New() with no Anthropic key should fail")
	} else if !strings.Contains(err.Error(), "set-anthropic-key") {
		t.Errorf("missing-key error should mention how to fix it, got: %v", err)
	}

	cfg.Anthropic.APIKey = "sk-ant-test"
	p, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if _, isAnthropic := p.(*AnthropicProvider); !isAnthropic {
		t.Errorf("New() returned %T, want *AnthropicProvider", p)
	}
}

func TestNewUnsupported(t *testing.T) {
	cfg := baseConfig()
	cfg.Provider = "carrier-pigeon"

	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() with unsupported provider should fail")
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Errorf("error should say unsupported provider, got: %v", err)
	}
	if !strings.Contains(err.Error(), "ollama, openai, anthropic") {
		t.Errorf("error should list supported providers, got: %v", err)
	}
}
