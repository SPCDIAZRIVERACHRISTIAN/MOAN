/*
Copyright © 2026 NAME HERE  christianda3@gmail.com

*/
package cmd

import (
	"fmt"
	"time"

	"github.com/SPCDIAZRIVERACHRISTIAN/moan/internal/review"
	"github.com/spf13/cobra"
)

var reviewCmd = &cobra.Command{
	Use:   "review",
	Short: "Collect repository changes for review",
	RunE: func(cmd *cobra.Command, args []string) error {
		stopLoader := startReviewLoader()

		result, err := review.Run()

		stopLoader()
		
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

		if result.ReviewContent != "" {
			fmt.Println()
			fmt.Println("AI REVIEW")
			fmt.Println("-----------------")
			fmt.Println(result.ReviewContent)
		} else {
			fmt.Println()
			fmt.Println("No model Response")
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(reviewCmd)
}

func startReviewLoader() func() {
	done := make(chan struct{})

	go func() {
		frames := []string{"[m   ]", "[mo  ]", "[moa ]", "[moan]", "[MOAN]"}
		i := 0

		for {
			select {
			case <-done:
				fmt.Print("\rMOAN> [done] review complete.              \n")
				return
			default:
				fmt.Printf("\rMOAN> %s chewing through your diff...", frames[i%len(frames)])
				i++
				time.Sleep(180 * time.Millisecond)
			}
		}
	}()

	return func() {
		close(done)
	}
}
