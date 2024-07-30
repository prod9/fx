package data

import (
	"os"
	"os/exec"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
	"github.com/spf13/cobra"
)

var psqlCmd = &cobra.Command{
	Use:   "psql",
	Short: "Starts a psql shell with the DATABASE_URL config",
	RunE:  runPSQLCmd,
}

func runPSQLCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("psql", &err)

	cfg := config.Configure()
	dbURL := config.Get(cfg, data.DatabaseURLConfig)

	// starts a psql shell with the DATABASE_URL config
	proc := exec.Command("psql", dbURL)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	return proc.Run()
}
