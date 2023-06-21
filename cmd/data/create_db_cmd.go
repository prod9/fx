package data

import (
	"log"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
	"github.com/spf13/cobra"
)

var createDBCmd = &cobra.Command{
	Use:   "create-db",
	Short: "Creates the database specified in the DATABASE_URL configuration.",
	Long:  "Creates the database specified by DATABASE_URL config, the user must have sufficient permissions.",
	RunE:  runCreateDBCmd,
}

func runCreateDBCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("create-db", &err)

	if err = data.CreateDB(config.Configure()); err != nil {
		return err
	} else {
		log.Println("database created.")
		return nil
	}
}
