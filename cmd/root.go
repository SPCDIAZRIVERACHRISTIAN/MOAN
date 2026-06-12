/*
Copyright © 2026 11b_shrink christrianda3@gmail.com
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Exit codes, so MOAN is scriptable.
const (
	exitOK         = 0
	exitRuntime    = 1 // setup/runtime error
	exitValidation = 2 // validation failed
	exitNoChanges  = 3 // no git changes
	exitProvider   = 4 // provider/API error
)

type exitCodeError struct {
	code int
	err  error
}

func (e *exitCodeError) Error() string { return e.err.Error() }
func (e *exitCodeError) Unwrap() error { return e.err }

func exitWithCode(code int, err error) error {
	return &exitCodeError{code: code, err: err}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:           "moan",
	Short:         "moan is a code reviewer orchestrator",
	Long:          "MOAN analyzes code changes and runs structured review passes for bugs, security, and architecture.",
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err == nil {
		return
	}

	fmt.Fprintf(os.Stderr, "Error: %s\n", err)

	var ec *exitCodeError
	if errors.As(err, &ec) {
		os.Exit(ec.code)
	}
	os.Exit(exitRuntime)
}
