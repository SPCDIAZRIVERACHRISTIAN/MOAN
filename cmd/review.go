/*
Copyright © 2026 NAME HERE  christianda3@gmail.com

*/
package cmd

import (
	"fmt"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/review"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Collect repository changes for review",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := review.Run()
		if err != nil {
			return err
		}

		if !result.Ready {
			fmt.Println("STATUS: NOT READY")
			fmt.Println("Repository is not ready for review.")
			return fmt.Errorf("review failed validation")
		}

		fmt.Println("STATUS: READY")
		fmt.Printf("Provider: %s\n", result.Provider)
		fmt.Printf("Model: %s\n", result.Model)
		fmt.Printf("Files changed: %d\n\n", len(result.Files))

		for _, file := range result.Files {
			fmt.Printf("- %s | +%d -%d\n", file.Path, file.Additions, file.Deletions)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}
