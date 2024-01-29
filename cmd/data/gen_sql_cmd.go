package data

import (
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/errutil"
	"github.com/go-jet/jet/v2/generator/postgres"
	_ "github.com/lib/pq" // still required by go-jet even though we're not using pgx
	"github.com/spf13/cobra"
)

var genSQLCmd = &cobra.Command{
	Use:   "gen-sql",
	Short: "Generate code for SQL builders",
	RunE:  runGenSQLCmd,
}

func runGenSQLCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("gen-sql", &err)

	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
		outdir = prompt.Str("output dir")
	)

	return postgres.GenerateDSN(
		config.Get(cfg, data.DatabaseURLConfig),
		"public",
		outdir,
	)
}
