package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/validate"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate repository state before review",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := validate.Run()
		if err != nil {
			return err
		}

		for _, check := range result.Checks {
			symbol := "✘"
			if check.Passed {
				symbol = "✔"
			}

			fmt.Printf("%s %s: %s\n", symbol, check.Name, check.Message)
		}

		if result.Valid {
			fmt.Println("\nSTATUS: READY")
			return nil
		}

		fmt.Println("\nSTATUS: NOT READY")
		return fmt.Errorf("validation failed")
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
