package data

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/fxlog"

	"github.com/spf13/cobra"
)

const upMigrationTemplate = `-- vim: filetype=SQL
CREATE TABLE dummy (
	id SERIAL PRIMARY KEY,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
`
const downMigrationTemplate = `-- vim: filetype=SQL
DROP TABLE dummy;
`

var newMigrationCmd = &cobra.Command{
	Use:     "new-migration (name) [subdir]",
	Aliases: []string{"new-migrate", "new"},
	Short:   "Creates a new migration file with timestamps and the given name",
	Run:     runNewMigrationCmd,
}

func runNewMigrationCmd(cmd *cobra.Command, args []string) {
	var (
		cfg    = config.Configure()
		prompt = prompts.New(cfg, args)
		dir    = config.Get(cfg, migrator.MigrationPathConfig)
	)

	name := prompt.Str("name of migration")

	// subdirectory selection is optional, defaults to top-level
	subdirs := []string{"."}
	entries, err := os.ReadDir(dir)
	if err != nil {
		fxlog.Fatalf("new-migration: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			subdirs = append(subdirs, entry.Name())
		}
	}

	dir = filepath.Join(dir, prompt.OptionalList("which subdirectory", subdirs[0], subdirs))

	// create the migration files
	uppath, downpath, err := migrator.MigrationPath(dir, name)
	if err != nil {
		fxlog.Fatalf("new-migration: %w", err)
	}

	fmt.Fprintln(os.Stdout, uppath)
	fmt.Fprintln(os.Stdout, downpath)
	if !prompt.YesNo("create these files") {
		fxlog.Fatalf("new-migration: aborted")
	}

	if err := ioutil.WriteFile(uppath, []byte(upMigrationTemplate), 0644); err != nil {
		fxlog.Fatalf("new-migration: %w", err)
	} else if err := ioutil.WriteFile(downpath, []byte(downMigrationTemplate), 0644); err != nil {
		fxlog.Fatalf("new-migration: %w", err)
	}

	editor := os.Getenv("EDITOR")
	if strings.TrimSpace(editor) == "" {
		editor = "/usr/bin/vi"
	}

	proc := exec.Command(editor, uppath, downpath)
	proc.Stdin = os.Stdin
	proc.Stdout = os.Stdout
	if err := proc.Run(); err != nil {
		fxlog.Fatalf("new-migration: %w", err)
	}
}
