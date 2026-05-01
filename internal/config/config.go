package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const FileName = ".moan.yml"

type Config struct {
	Provider     string `yaml:"provider"`
	Model        string `yaml:"model"`
	BaseURL      string `yaml:"base_url"`
	APIToken     string `yaml:"api_token"`
	SystemPrompt string `yaml:"system_prompt"`
}

func Load() (Config, error) {
	root, err := findRepoRoot()
	if err != nil {
		return Config{}, fmt.Errorf("find repo root: %w", err)
	}

	configPath := filepath.Join(root, FileName)

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("config file not found: %s", configPath)
		}
		return Config{}, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config file: %w", err)
	}

	applyDefaults(&cfg)

	return cfg, nil
}

func Save(cfg Config) error {
	if err := Validate(cfg); err != nil {
		return err
	}

	root, err := findRepoRoot()
	if err != nil {
		return fmt.Errorf("find repo root: %w", err)
	}

	configPath := filepath.Join(root, FileName)

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

func Validate(cfg Config) error {
	if cfg.Provider == "" {
		return fmt.Errorf("provider is required")
	}

	if cfg.Model == "" {
		return fmt.Errorf("model is required")
	}

	switch cfg.Provider {
	case "ollama":
		if cfg.BaseURL == "" {
			return fmt.Errorf("base_url is required for ollama")
		}
	default:
		return fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}

	return nil
}

func applyDefaults(cfg *Config) {
	if cfg.Provider == "ollama" && cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:11434"
	}

	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = "You are MOAN, a strict code reviewer focused on bugs, architecture, maintainability, and security."
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	for {
		gitPath := filepath.Join(dir, ".git")
		if _, err := os.Stat(gitPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find repository root")
		}

		dir = parent
	}
}
