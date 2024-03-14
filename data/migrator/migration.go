package migrator

import (
	"path/filepath"
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
