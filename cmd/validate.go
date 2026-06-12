/*
Copyright © 2026 11b_shrink christianda3@gmail.com
*/
package cmd

import (
	"fmt"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/config"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/validate"
	"github.com/spf13/cobra"
)

var (
	validateProvider string
	validateModel    string
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate repository, config, and provider before review",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load()
		if err != nil {
			return exitWithCode(exitRuntime, err)
		}

		if validateProvider != "" {
			cfg.Provider = validateProvider
		}
		if validateModel != "" {
			cfg.Model = validateModel
		}

		result := validate.Run(cfg)

		fmt.Println("MOAN VALIDATION")
		fmt.Println("---------------")
		for _, item := range result.Items {
			fmt.Printf("%s: %s\n", item.Label, item.Value)
		}

		if result.Valid {
			fmt.Println("\nSTATUS: READY")
			return nil
		}

		fmt.Println("\nSTATUS: NOT READY")
		return exitWithCode(exitValidation, fmt.Errorf("validation failed"))
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)

	validateCmd.Flags().StringVar(&validateProvider, "provider", "", "override provider for this run (ollama, openai, anthropic)")
	validateCmd.Flags().StringVar(&validateModel, "model", "", "override model for this run")
}
