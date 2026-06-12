package config

import (
	"os"
	"path/filepath"
	"testing"
)

// clearEnv blanks every env var the config system reads so tests are
// isolated from the host environment.
func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"MOAN_CONFIG", "MOAN_PROVIDER", "MOAN_MODEL",
		"OLLAMA_BASE_URL", "OPENAI_API_KEY", "OPENAI_BASE_URL",
		"ANTHROPIC_API_KEY", "ANTHROPIC_BASE_URL",
	} {
		t.Setenv(key, "")
	}
}

func pointConfigAt(t *testing.T, path string) {
	t.Helper()
	t.Setenv("MOAN_CONFIG", path)
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	pointConfigAt(t, filepath.Join(t.TempDir(), "config.yaml"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Provider != ProviderOllama {
		t.Errorf("Provider = %q, want %q", cfg.Provider, ProviderOllama)
	}
	if cfg.TimeoutSeconds != DefaultTimeoutSeconds {
		t.Errorf("TimeoutSeconds = %d, want %d", cfg.TimeoutSeconds, DefaultTimeoutSeconds)
	}
	if cfg.MaxDiffChars != DefaultMaxDiffChars {
		t.Errorf("MaxDiffChars = %d, want %d", cfg.MaxDiffChars, DefaultMaxDiffChars)
	}
	if cfg.MaxFiles != DefaultMaxFiles {
		t.Errorf("MaxFiles = %d, want %d", cfg.MaxFiles, DefaultMaxFiles)
	}
	if cfg.Ollama.BaseURL != DefaultOllamaBaseURL {
		t.Errorf("Ollama.BaseURL = %q, want %q", cfg.Ollama.BaseURL, DefaultOllamaBaseURL)
	}
	if cfg.OpenAI.BaseURL != DefaultOpenAIBaseURL {
		t.Errorf("OpenAI.BaseURL = %q, want %q", cfg.OpenAI.BaseURL, DefaultOpenAIBaseURL)
	}
	if cfg.Anthropic.BaseURL != DefaultAnthropicBaseURL {
		t.Errorf("Anthropic.BaseURL = %q, want %q", cfg.Anthropic.BaseURL, DefaultAnthropicBaseURL)
	}
	if got := cfg.ResolvedModel(); got != DefaultOllamaModel {
		t.Errorf("ResolvedModel() = %q, want %q", got, DefaultOllamaModel)
	}
}

func TestLoadFromFile(t *testing.T) {
	clearEnv(t)

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `provider: openai
openai:
  api_key: sk-file-key
  model: gpt-test
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	pointConfigAt(t, path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Provider != ProviderOpenAI {
		t.Errorf("Provider = %q, want openai", cfg.Provider)
	}
	if cfg.OpenAI.APIKey != "sk-file-key" {
		t.Errorf("OpenAI.APIKey = %q, want sk-file-key", cfg.OpenAI.APIKey)
	}
	if got := cfg.ResolvedModel(); got != "gpt-test" {
		t.Errorf("ResolvedModel() = %q, want gpt-test", got)
	}
}

func TestEnvOverridesFile(t *testing.T) {
	clearEnv(t)

	path := filepath.Join(t.TempDir(), "config.yaml")
	content := `provider: ollama
openai:
  api_key: sk-file-key
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	pointConfigAt(t, path)

	t.Setenv("MOAN_PROVIDER", "anthropic")
	t.Setenv("MOAN_MODEL", "claude-test")
	t.Setenv("OPENAI_API_KEY", "sk-env-key")
	t.Setenv("ANTHROPIC_API_KEY", "sk-ant-env")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Provider != ProviderAnthropic {
		t.Errorf("Provider = %q, want anthropic (env should override file)", cfg.Provider)
	}
	if cfg.OpenAI.APIKey != "sk-env-key" {
		t.Errorf("OpenAI.APIKey = %q, want sk-env-key (env should override file)", cfg.OpenAI.APIKey)
	}
	if cfg.Anthropic.APIKey != "sk-ant-env" {
		t.Errorf("Anthropic.APIKey = %q, want sk-ant-env", cfg.Anthropic.APIKey)
	}
	if got := cfg.ResolvedModel(); got != "claude-test" {
		t.Errorf("ResolvedModel() = %q, want claude-test (MOAN_MODEL should win)", got)
	}
}

func TestLoadFileWithDefaultsIgnoresEnv(t *testing.T) {
	clearEnv(t)
	pointConfigAt(t, filepath.Join(t.TempDir(), "config.yaml"))
	t.Setenv("OPENAI_API_KEY", "sk-env-key")

	cfg, err := LoadFileWithDefaults()
	if err != nil {
		t.Fatalf("LoadFileWithDefaults() error: %v", err)
	}

	if cfg.OpenAI.APIKey != "" {
		t.Errorf("OpenAI.APIKey = %q, want empty (env must not leak into saved config)", cfg.OpenAI.APIKey)
	}
}

func TestSaveAndReload(t *testing.T) {
	clearEnv(t)
	pointConfigAt(t, filepath.Join(t.TempDir(), "sub", "config.yaml"))

	cfg, err := LoadFileWithDefaults()
	if err != nil {
		t.Fatal(err)
	}
	cfg.Provider = ProviderAnthropic
	cfg.Anthropic.APIKey = "sk-ant-saved"

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load() after Save error: %v", err)
	}

	if reloaded.Provider != ProviderAnthropic {
		t.Errorf("Provider = %q, want anthropic", reloaded.Provider)
	}
	if reloaded.Anthropic.APIKey != "sk-ant-saved" {
		t.Errorf("Anthropic.APIKey = %q, want sk-ant-saved", reloaded.Anthropic.APIKey)
	}
}

func TestValidate(t *testing.T) {
	clearEnv(t)
	pointConfigAt(t, filepath.Join(t.TempDir(), "config.yaml"))

	cfg, _ := Load()
	if err := Validate(cfg); err != nil {
		t.Errorf("default ollama config should validate, got: %v", err)
	}

	cfg.Provider = "carrier-pigeon"
	if err := Validate(cfg); err == nil {
		t.Error("unsupported provider should fail validation")
	}

	cfg.Provider = ProviderOpenAI
	cfg.OpenAI.APIKey = ""
	if err := Validate(cfg); err == nil {
		t.Error("openai without API key should fail validation")
	}

	cfg.OpenAI.APIKey = "sk-test"
	if err := Validate(cfg); err != nil {
		t.Errorf("openai with API key should validate, got: %v", err)
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", "(not set)"},
		{"short", "****"},
		{"sk-abcdefghijklmnop", "sk-...mnop"},
		{"sk-ant-abcdefghwxyz", "sk-ant-...wxyz"},
		{"nodashesatall1234", "noda...1234"},
	}

	for _, tt := range tests {
		if got := MaskKey(tt.in); got != tt.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
