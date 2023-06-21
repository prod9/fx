package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/errutil"

	"github.com/spf13/cobra"
)

const upMigrationTemplate = `-- vim: filetype=SQL
CREATE TABLE dummy (
	id TEXT PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
const downMigrationTemplate = `-- vim: filetype=SQL
DROP TABLE dummy;
`

var newMigrationCmd = &cobra.Command{
	Use:     "new-migration (name)",
	Aliases: []string{"new-migrate"},
	Short:   "Creates a new migration file with timestamps and the given name",
	RunE:    runNewMigrationCmd,
}

func runNewMigrationCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("new-migration", &err)

	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
		name   = prompt.Str("name of migration")
		dir    = config.Get(cfg, data.MigrationPathConfig)
	)

	uppath, downpath, err := migrator.MigrationPath(dir, name)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stdout, uppath)
	fmt.Fprintln(os.Stdout, downpath)
	if !prompt.YesNo("create these files") {
		log.Fatalln("aborted")
		return
	}

	if err := ioutil.WriteFile(uppath, []byte(upMigrationTemplate), 0644); err != nil {
		return err
	} else if err := ioutil.WriteFile(downpath, []byte(downMigrationTemplate), 0644); err != nil {
		return err
	}

	editor := os.Getenv("EDITOR")
	if strings.TrimSpace(editor) == "" {
		editor = "/usr/bin/vi"
	}

	proc := exec.Command(editor, uppath, downpath)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	return proc.Run()
}
