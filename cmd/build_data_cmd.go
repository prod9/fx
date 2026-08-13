package cmd

import (
	"fx.prodigy9.co/cmd/data"
	"fx.prodigy9.co/data/migrator"
	"github.com/spf13/cobra"
)

// BuildDataCommand builds the `data` command group. Explicit migration sources
// (typically an app tree's collected embedded migrations) are threaded into the
// migration-reading subcommands; with none given, the commands fall back to the
// auto-detected sources alone.
func BuildDataCommand(srcs ...migrator.Source) *cobra.Command {
	root := &cobra.Command{
		Use:   "data",
		Short: "Work with databases",
	}
	root.AddCommand(data.Commands(srcs...)...)
	return root
}
