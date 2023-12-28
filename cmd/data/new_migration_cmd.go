package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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
		dir    = config.Get(cfg, data.MigrationPathConfig)
	)

	// we don't normally add migraitons to top-level, so let user pick a folder first
	subdirs := []string{"."}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subdirs = append(subdirs, entry.Name())
		}
	}

	dir = filepath.Join(dir, prompt.List("which subdirectory", subdirs[0], subdirs))
	name := prompt.Str("name of migration")

	// create the migration files
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
