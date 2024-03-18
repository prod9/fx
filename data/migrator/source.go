package migrator

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fx.prodigy9.co/config"
)

var (
	MigrationPathConfig = config.StrDef("DATABASE_MIGRATIONS", ".")
	ErrNoMigrations     = errors.New("migrator: no migrations found")

	embeddedMigrations fs.FS
)

func IsNoMigrations(err error) bool {
	return errors.Is(err, ErrNoMigrations)
}

type Source func() ([]Migration, error)

func FromDir(dir string) Source { return func() ([]Migration, error) { return loadMigrations(dir) } }
func FromFS(fsys fs.FS) Source  { return func() ([]Migration, error) { return loadMigrationsFS(fsys) } }
func FromConfig(src *config.Source) Source {
	return func() ([]Migration, error) { return loadMigrations(config.Get(src, MigrationPathConfig)) }
}
func FromAuto(cfg *config.Source) Source {
	return func() ([]Migration, error) { return LoadAuto(cfg) }
}

// Embed provide a way for compiled Go programs to embed migration files directly into the
// binary so that we don't have to take care of the logistics of putting them on prod
// servers, containers, CLIs etc.
//
// Create an embedded embed.FS variable in your application's main.go file like so:
//
//	//go:embed migrations
//	var migrationsFS embed.FS
//
//	func init() {
//		migrator.Embed(migrationsFS)
//	}
func Embed(fsys embed.FS) {
	embeddedMigrations = fsys
}

// LoadAuto checks a few different hard-coded sources automagically for migrations in the
// following order:
//
// 1. Path configured via DATABASE_MIGRATIONS env var.
// 2. Local working directory.
// 3. Embedded migrations.
//
// The intended use case is that developer will be working with local files during
// development and will use Embed to embed the migrations into the binary for production
// deployments.
//
// In case a custom workflow is required, set DATABASE_MIGRATIONS env var configuration
// to the desired path since it will always take precedence over everything else.
func LoadAuto(cfg *config.Source) ([]Migration, error) {
	migPath, ok := config.GetOK(cfg, MigrationPathConfig)
	if ok {
		// if the env var is set, but there are no migrations, we let it errors because it's
		// likely a misconfiguration.
		return Load(FromDir(migPath))
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("migrator: %w", err)
	}

	migrations, err := Load(FromDir(wd))
	if err != nil && !IsNoMigrations(err) {
		return nil, fmt.Errorf("migrator: %w", err)
	}

	if len(migrations) > 0 {
		return migrations, nil
	}

	// last chance, look for embedded migrations
	if embeddedMigrations == nil {
		return nil, ErrNoMigrations
	} else {
		return Load(FromFS(embeddedMigrations))
	}
}

func Load(src Source) ([]Migration, error) {
	if migrations, err := src(); err != nil {
		return nil, err
	} else if len(migrations) == 0 {
		return nil, ErrNoMigrations
	} else {
		return migrations, nil
	}
}

func loadMigrationsFS(fsys fs.FS) (result []Migration, err error) {
	err = fs.WalkDir(fsys, ".", func(path string, info fs.DirEntry, err error) error {
		if info.IsDir() {
			if strings.ToLower(info.Name()) == ".git" ||
				strings.ToLower(info.Name()) == "node_modules" ||
				strings.ToLower(info.Name()) == "vendor" ||
				strings.ToLower(info.Name()) == "examples" {
				return fs.SkipDir
			} else {
				return nil
			}
		}

		matched, err := filepath.Match("*"+UpExt, info.Name())
		if err != nil {
			return err
		} else if !matched {
			return nil
		}

		migration, err := loadOneMigration(fsys, path)
		if err != nil {
			return err
		}

		result = append(result, migration)
		return nil
	})

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return
}

func loadMigrations(dir string) (result []Migration, err error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("migrator: %w", err)
	}

	fsys := os.DirFS(abs)
	return loadMigrationsFS(fsys)
}

func loadOneMigration(fsys fs.FS, path string) (Migration, error) {
	var (
		basename  = filepath.Base(path)
		dirname   = filepath.Dir(path)
		downfile  = path[:len(path)-len(UpExt)] + DownExt
		migration = Migration{
			Name: basename[:len(basename)-len(UpExt)],
		}
	)

	if _, err := fs.Stat(fsys, downfile); errors.Is(err, fs.ErrNotExist) {
		return Migration{}, fmt.Errorf("migrator: missing down migration: %s", downfile)
	}
	if bytes, err := fs.ReadFile(fsys, path); err != nil {
		return Migration{}, fmt.Errorf("migrator: %w", err)
	} else {
		migration.UpSQL = string(bytes)
	}

	if bytes, err := fs.ReadFile(fsys, downfile); err != nil {
		return Migration{}, fmt.Errorf("migrator: %w", err)
	} else {
		migration.DownSQL = string(bytes)
	}

	migration.Dir = dirname
	return migration, nil
}
