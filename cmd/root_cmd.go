package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func BuildRootCommand(desc string, cmds ...*cobra.Command) *cobra.Command {
	desc = strings.TrimSpace(desc)
	if desc == "" {
		desc = "PRODIGY9 FX Application"
	}

	rootCmd := &cobra.Command{Use: "app", Short: desc}
	rootCmd.AddCommand(cmds...)
	return rootCmd
}
