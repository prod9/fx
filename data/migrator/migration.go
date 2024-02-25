package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/chakrit/gendiff"
	"github.com/gobuffalo/flect"
	"github.com/jmoiron/sqlx"
)

type Migration struct {
	Name    string `db:"name"`
	Dir     string `db:"-"`
	UpSQL   string `db:"up_sql"`
	DownSQL string `db:"down_sql"`
}

type migrationDiff struct {
	left  []Migration
	right []Migration
}

var _ gendiff.Interface = migrationDiff{}

func (d migrationDiff) LeftLen() int        { return len(d.left) }
func (d migrationDiff) RightLen() int       { return len(d.right) }
func (d migrationDiff) Equal(l, r int) bool { return d.left[l].Name == d.right[r].Name }

func MigrationPath(dir, name string) (string, string, error) {
	name = time.Now().Format("200601021504") + "_" + flect.Underscore(name)
	upname, downname := name+UpExt, name+DownExt

	uppath, err := filepath.Abs(filepath.Join(dir, upname))
	if err != nil {
		return "", "", err
	}
	downpath, err := filepath.Abs(filepath.Join(dir, downname))
	if err != nil {
		return "", "", err
	}

	return uppath, downpath, nil
}

func RecoverMigrations(db *sqlx.DB) (migrations []Migration, err error) {
	err = db.Select(&migrations, ListMigrationsSQL)
	return
}

func LoadMigrations(dir string) (result []Migration, err error) {
	err = filepath.WalkDir(dir, func(path string, info os.DirEntry, err error) error {
		if info.IsDir() {
			if strings.ToLower(info.Name()) == ".git" ||
				strings.ToLower(info.Name()) == "node_modules" ||
				strings.ToLower(info.Name()) == "vendor" ||
				strings.ToLower(info.Name()) == "examples" {
				return filepath.SkipDir
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

		migration, err := loadOneMigration(path)
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

func loadOneMigration(path string) (Migration, error) {
	var (
		basename  = filepath.Base(path)
		dirname   = filepath.Dir(path)
		downfile  = path[:len(path)-len(UpExt)] + DownExt
		migration = Migration{
			Name: basename[:len(basename)-len(UpExt)],
		}
	)

	if _, err := os.Stat(downfile); os.IsNotExist(err) {
		return Migration{}, fmt.Errorf("migrator: missing down migration: %s", downfile)
	}
	if bytes, err := os.ReadFile(path); err != nil {
		return Migration{}, fmt.Errorf("migrator: %w", err)
	} else {
		migration.UpSQL = string(bytes)
	}

	if bytes, err := os.ReadFile(downfile); err != nil {
		return Migration{}, fmt.Errorf("migrator: %w", err)
	} else {
		migration.DownSQL = string(bytes)
	}

	migration.Dir = dirname
	return migration, nil
}
