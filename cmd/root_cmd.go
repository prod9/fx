package cmd

import (
	"strings"

	"fx.prodigy9.co/cmd/data"
	"github.com/spf13/cobra"
)

func BuildRootCommand(desc string, extraCmds ...*cobra.Command) *cobra.Command {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		desc = "PRODIGY9 Application"
	}

	rootCmd := &cobra.Command{Use: "app", Short: desc}
	rootCmd.AddCommand(
		PrintConfigCmd,
		testEmailCmd,
		data.Cmd,
	)

	rootCmd.AddCommand(extraCmds...)
	return rootCmd
}
