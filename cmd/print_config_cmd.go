package cmd

import (
	"fmt"
	"os"

	"fx.prodigy9.co/config"
	"github.com/spf13/cobra"
)

var PrintConfigCmd = &cobra.Command{
	Use:   "print-config",
	Short: "Prints current effective configuration.",
	RunE:  runPrintConfigCmd,
}

func runPrintConfigCmd(cmd *cobra.Command, args []string) error {
	cfg := config.Configure()
	for _, v := range cfg.Vars() {
		fmt.Fprintln(os.Stdout, v.Name(), "=", config.GetAny(cfg, v))
	}
	return nil
}
