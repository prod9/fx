package data

import (
	"os"
	"os/exec"

	"fx.prodigy9.co/cmd/cmdutil"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Starts a psql shell with the DATABASE_URL config",
	Run:   runPSQLCmd,
}

func runPSQLCmd(cmd *cobra.Command, args []string) {
	_, cfg := cmdutil.NewBasicContext()
	dbURL := config.Get(cfg, data.DatabaseURLConfig)

	// starts a psql shell with the DATABASE_URL config
	proc := exec.Command("psql", dbURL)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	if err := proc.Run(); err != nil {
		fxlog.Fatalf("psql: %w", err)
	}
}
