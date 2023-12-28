package cmd

import (
	"fmt"
	"os"
	"strings"

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
	if len(args) == 0 {
		for _, v := range cfg.Vars() {
			fmt.Fprintln(os.Stdout, v.Name(), "=", config.GetAny(cfg, v))
		}

	} else {
		// print only var in arg, and only its content. useful for scripting/programmatically
		// extracting effective config value from the app
		//
		// for example, you can invoke psql to the configured database by doing:
		//
		//    psql "$(./app print-config DATABASE_URL)"
		//
		for _, v := range cfg.Vars() {
			if strings.EqualFold(v.Name(), args[0]) {
				fmt.Fprintln(os.Stdout, config.GetAny(cfg, v))
				break
			}
		}
	}

	return nil
}
