/*
Copyright © 2026 11b_shrink christianda3@gmail.com
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage MOAN configuration",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Create a default config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}

		if config.Exists() {
			fmt.Printf("Config already exists at %s\n", path)
			return nil
		}

		cfg, err := config.LoadFileWithDefaults()
		if err != nil {
			return err
		}

		if err := config.Save(cfg); err != nil {
			return err
		}

		fmt.Printf("Config created at %s\n", path)
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the current configuration (API keys masked)",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.Path()
		if err != nil {
			return err
		}

		cfg, err := config.Load()
		if err != nil {
			return err
		}

		masked := cfg
		masked.OpenAI.APIKey = config.MaskKey(cfg.OpenAI.APIKey)
		masked.Anthropic.APIKey = config.MaskKey(cfg.Anthropic.APIKey)

		data, err := yaml.Marshal(&masked)
		if err != nil {
			return fmt.Errorf("marshal config: %w", err)
		}

		fmt.Printf("Config file: %s (exists: %v)\n", path, config.Exists())
		fmt.Println("Effective config (file + env + defaults):")
		fmt.Println()
		fmt.Print(string(data))
		return nil
	},
}

var configSetProviderCmd = &cobra.Command{
	Use:   "set-provider <ollama|openai|anthropic>",
	Short: "Set the default provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		provider := strings.ToLower(args[0])
		if !config.SupportedProvider(provider) {
			return fmt.Errorf("unsupported provider: %q (supported: ollama, openai, anthropic)", provider)
		}

		return updateConfig(func(cfg *config.Config) {
			cfg.Provider = provider
		}, "provider set to "+provider)
	},
}

var configSetModelCmd = &cobra.Command{
	Use:   "set-model <model>",
	Short: "Set the model for the currently selected provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		model := args[0]

		return updateConfig(func(cfg *config.Config) {
			switch cfg.Provider {
			case config.ProviderOpenAI:
				cfg.OpenAI.Model = model
			case config.ProviderAnthropic:
				cfg.Anthropic.Model = model
			default:
				cfg.Ollama.Model = model
			}
		}, "model set to "+model)
	},
}

var configSetOpenAIKeyCmd = &cobra.Command{
	Use:   "set-openai-key <key>",
	Short: "Set the OpenAI API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateConfig(func(cfg *config.Config) {
			cfg.OpenAI.APIKey = args[0]
		}, "OpenAI API key saved ("+config.MaskKey(args[0])+")")
	},
}

var configSetAnthropicKeyCmd = &cobra.Command{
	Use:   "set-anthropic-key <key>",
	Short: "Set the Anthropic API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateConfig(func(cfg *config.Config) {
			cfg.Anthropic.APIKey = args[0]
		}, "Anthropic API key saved ("+config.MaskKey(args[0])+")")
	},
}

var configSetOllamaURLCmd = &cobra.Command{
	Use:   "set-ollama-url <url>",
	Short: "Set the Ollama base URL",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return updateConfig(func(cfg *config.Config) {
			cfg.Ollama.BaseURL = args[0]
		}, "Ollama base URL set to "+args[0])
	},
}

var configSetTimeoutCmd = &cobra.Command{
	Use:   "set-timeout <seconds>",
	Short: "Set the request timeout in seconds",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		seconds, err := strconv.Atoi(args[0])
		if err != nil || seconds <= 0 {
			return fmt.Errorf("timeout must be a positive number of seconds, got %q", args[0])
		}

		return updateConfig(func(cfg *config.Config) {
			cfg.TimeoutSeconds = seconds
		}, fmt.Sprintf("timeout set to %ds", seconds))
	},
}

var configSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Interactive configuration setup",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadFileWithDefaults()
		if err != nil {
			return err
		}

		reader := bufio.NewReader(os.Stdin)

		provider := ask(reader, "Provider (ollama/openai/anthropic)", cfg.Provider)
		provider = strings.ToLower(provider)
		if !config.SupportedProvider(provider) {
			return fmt.Errorf("unsupported provider: %q (supported: ollama, openai, anthropic)", provider)
		}
		cfg.Provider = provider

		switch provider {
		case config.ProviderOllama:
			cfg.Ollama.Model = ask(reader, "Model", cfg.Ollama.Model)
			cfg.Ollama.BaseURL = ask(reader, "Base URL", cfg.Ollama.BaseURL)
		case config.ProviderOpenAI:
			cfg.OpenAI.Model = ask(reader, "Model", cfg.OpenAI.Model)
			if key := ask(reader, "OpenAI API key ("+config.MaskKey(cfg.OpenAI.APIKey)+", enter to keep)", ""); key != "" {
				cfg.OpenAI.APIKey = key
			}
			cfg.OpenAI.BaseURL = ask(reader, "Base URL", cfg.OpenAI.BaseURL)
		case config.ProviderAnthropic:
			cfg.Anthropic.Model = ask(reader, "Model", cfg.Anthropic.Model)
			if key := ask(reader, "Anthropic API key ("+config.MaskKey(cfg.Anthropic.APIKey)+", enter to keep)", ""); key != "" {
				cfg.Anthropic.APIKey = key
			}
			cfg.Anthropic.BaseURL = ask(reader, "Base URL", cfg.Anthropic.BaseURL)
		}

		timeoutStr := ask(reader, "Timeout seconds", strconv.Itoa(cfg.TimeoutSeconds))
		seconds, err := strconv.Atoi(timeoutStr)
		if err != nil || seconds <= 0 {
			return fmt.Errorf("timeout must be a positive number of seconds, got %q", timeoutStr)
		}
		cfg.TimeoutSeconds = seconds

		if err := config.Save(cfg); err != nil {
			return err
		}

		path, _ := config.Path()
		fmt.Printf("\nConfig saved to %s\n", path)
		return nil
	},
}

func ask(reader *bufio.Reader, label, defaultValue string) string {
	if defaultValue != "" {
		fmt.Printf("%s [%s]: ", label, defaultValue)
	} else {
		fmt.Printf("%s: ", label)
	}

	line, err := reader.ReadString('\n')
	if err != nil {
		return defaultValue
	}

	line = strings.TrimSpace(line)
	if line == "" {
		return defaultValue
	}
	return line
}

// updateConfig loads the config file (without env overrides, so env values
// never get written to disk), applies the change, and saves.
func updateConfig(mutate func(*config.Config), successMsg string) error {
	cfg, err := config.LoadFileWithDefaults()
	if err != nil {
		return err
	}

	mutate(&cfg)

	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Println(successMsg)
	return nil
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetupCmd)
	configCmd.AddCommand(configSetProviderCmd)
	configCmd.AddCommand(configSetModelCmd)
	configCmd.AddCommand(configSetOpenAIKeyCmd)
	configCmd.AddCommand(configSetAnthropicKeyCmd)
	configCmd.AddCommand(configSetOllamaURLCmd)
	configCmd.AddCommand(configSetTimeoutCmd)
}
