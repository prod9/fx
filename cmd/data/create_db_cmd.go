package data

import (
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/fxlog"
	"github.com/spf13/cobra"
)

var createDBCmd = &cobra.Command{
	Use:   "create-db",
	Short: "Creates the database specified in the DATABASE_URL configuration.",
	Long:  "Creates the database specified by DATABASE_URL config, the user must have sufficient permissions.",
	Run:   runCreateDBCmd,
}

func runCreateDBCmd(cmd *cobra.Command, args []string) {
	cfg := config.Configure()
	if err := data.CreateDB(cfg); err != nil {
		fxlog.Fatalf("create-db: %w", err)
	} else {
		fxlog.Log("database created")
	}
}
