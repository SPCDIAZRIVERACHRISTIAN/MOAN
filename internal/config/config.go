package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	ProviderOllama    = "ollama"
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"

	DefaultProvider         = ProviderOllama
	DefaultOllamaBaseURL    = "http://localhost:11434"
	DefaultOllamaModel      = "qwen2.5-coder:7b"
	DefaultOpenAIBaseURL    = "https://api.openai.com/v1"
	DefaultOpenAIModel      = "gpt-4o-mini"
	DefaultAnthropicBaseURL = "https://api.anthropic.com"
	DefaultAnthropicModel   = "claude-3-5-haiku-latest"
	DefaultTimeoutSeconds   = 120
	DefaultMaxDiffChars     = 60000
	DefaultMaxFiles         = 30
	DefaultSystemPrompt     = "You are MOAN, a strict code reviewer focused on bugs, architecture, maintainability, and security."
)

type OllamaConfig struct {
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type OpenAIConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type AnthropicConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Model   string `yaml:"model"`
}

type Config struct {
	Provider       string `yaml:"provider"`
	Model          string `yaml:"model"`
	SystemPrompt   string `yaml:"system_prompt"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
	MaxDiffChars   int    `yaml:"max_diff_chars"`
	MaxFiles       int    `yaml:"max_files"`

	Ollama    OllamaConfig    `yaml:"ollama"`
	OpenAI    OpenAIConfig    `yaml:"openai"`
	Anthropic AnthropicConfig `yaml:"anthropic"`
}

// Path returns the config file location: $MOAN_CONFIG if set, otherwise
// the platform user config directory (e.g. ~/.config/moan/config.yaml).
func Path() (string, error) {
	if p := os.Getenv("MOAN_CONFIG"); p != "" {
		return p, nil
	}

	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve user config directory: %w", err)
	}

	return filepath.Join(dir, "moan", "config.yaml"), nil
}

func Exists() bool {
	path, err := Path()
	if err != nil {
		return false
	}

	_, err = os.Stat(path)
	return err == nil
}

// Load reads the config file (if present), applies environment variable
// overrides, then fills remaining gaps with safe defaults. A missing config
// file is not an error.
//
// Priority (highest first): CLI flags (applied by the cmd layer on the
// returned struct), environment variables, config file, defaults.
func Load() (Config, error) {
	cfg, err := loadFile()
	if err != nil {
		return Config{}, err
	}

	applyEnv(&cfg)
	applyDefaults(&cfg)

	return cfg, nil
}

// LoadFileWithDefaults reads only the config file plus defaults, skipping
// environment overrides. The `moan config set-*` commands use this so that
// environment values never get baked into the saved file.
func LoadFileWithDefaults() (Config, error) {
	cfg, err := loadFile()
	if err != nil {
		return Config{}, err
	}

	applyDefaults(&cfg)
	return cfg, nil
}

func loadFile() (Config, error) {
	var cfg Config

	path, err := Path()
	if err != nil {
		return Config{}, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return Config{}, fmt.Errorf("read config file %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file %s: %w", path, err)
	}

	return cfg, nil
}

func Save(cfg Config) error {
	path, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func applyEnv(cfg *Config) {
	if v := os.Getenv("MOAN_PROVIDER"); v != "" {
		cfg.Provider = v
	}
	if v := os.Getenv("MOAN_MODEL"); v != "" {
		cfg.Model = v
	}
	if v := os.Getenv("OLLAMA_BASE_URL"); v != "" {
		cfg.Ollama.BaseURL = v
	}
	if v := os.Getenv("OPENAI_API_KEY"); v != "" {
		cfg.OpenAI.APIKey = v
	}
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
		cfg.OpenAI.BaseURL = v
	}
	if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
		cfg.Anthropic.APIKey = v
	}
	if v := os.Getenv("ANTHROPIC_BASE_URL"); v != "" {
		cfg.Anthropic.BaseURL = v
	}
}

func applyDefaults(cfg *Config) {
	if cfg.Provider == "" {
		cfg.Provider = DefaultProvider
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = DefaultSystemPrompt
	}
	if cfg.TimeoutSeconds <= 0 {
		cfg.TimeoutSeconds = DefaultTimeoutSeconds
	}
	if cfg.MaxDiffChars <= 0 {
		cfg.MaxDiffChars = DefaultMaxDiffChars
	}
	if cfg.MaxFiles <= 0 {
		cfg.MaxFiles = DefaultMaxFiles
	}

	if cfg.Ollama.BaseURL == "" {
		cfg.Ollama.BaseURL = DefaultOllamaBaseURL
	}
	if cfg.Ollama.Model == "" {
		cfg.Ollama.Model = DefaultOllamaModel
	}

	if cfg.OpenAI.BaseURL == "" {
		cfg.OpenAI.BaseURL = DefaultOpenAIBaseURL
	}
	if cfg.OpenAI.Model == "" {
		cfg.OpenAI.Model = DefaultOpenAIModel
	}

	if cfg.Anthropic.BaseURL == "" {
		cfg.Anthropic.BaseURL = DefaultAnthropicBaseURL
	}
	if cfg.Anthropic.Model == "" {
		cfg.Anthropic.Model = DefaultAnthropicModel
	}
}

// ResolvedModel returns the model for the active provider. A top-level
// Model (set via --model or MOAN_MODEL) wins; otherwise the per-provider
// model from the config file or defaults is used.
func (c Config) ResolvedModel() string {
	if c.Model != "" {
		return c.Model
	}

	switch c.Provider {
	case ProviderOllama:
		return c.Ollama.Model
	case ProviderOpenAI:
		return c.OpenAI.Model
	case ProviderAnthropic:
		return c.Anthropic.Model
	default:
		return ""
	}
}

func SupportedProvider(name string) bool {
	switch name {
	case ProviderOllama, ProviderOpenAI, ProviderAnthropic:
		return true
	default:
		return false
	}
}

func Validate(cfg Config) error {
	if !SupportedProvider(cfg.Provider) {
		return fmt.Errorf("unsupported provider: %q (supported: ollama, openai, anthropic)", cfg.Provider)
	}

	if cfg.ResolvedModel() == "" {
		return fmt.Errorf("model is required")
	}

	if cfg.TimeoutSeconds <= 0 {
		return fmt.Errorf("timeout_seconds must be positive")
	}

	switch cfg.Provider {
	case ProviderOllama:
		if cfg.Ollama.BaseURL == "" {
			return fmt.Errorf("ollama.base_url is required")
		}
	case ProviderOpenAI:
		if cfg.OpenAI.APIKey == "" {
			return fmt.Errorf("OpenAI API key is missing.\nSet it with:\n  moan config set-openai-key <key>\nor:\n  export OPENAI_API_KEY=<key>")
		}
	case ProviderAnthropic:
		if cfg.Anthropic.APIKey == "" {
			return fmt.Errorf("Anthropic API key is missing.\nSet it with:\n  moan config set-anthropic-key <key>\nor:\n  export ANTHROPIC_API_KEY=<key>")
		}
	}

	return nil
}

// MaskKey hides the middle of an API key for display: "sk-ant-...wxyz".
// It keeps a recognizable dashed prefix when one exists in the first 8
// characters, plus the last 4 characters.
func MaskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "****"
	}

	head := key[:8]
	prefix := key[:4]
	if i := strings.LastIndex(head, "-"); i >= 0 {
		prefix = key[:i+1]
	}

	return prefix + "..." + key[len(key)-4:]
}
