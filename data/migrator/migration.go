package migrator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/chakrit/gendiff"
	"github.com/gobuffalo/flect"
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

func LoadMigrations(dir string) (result []Migration, err error) {
	files, err := filepath.Glob(filepath.Clean(dir) + "/*" + UpExt)
	if err != nil {
		return nil, err
	}
	nestedFiles, err := filepath.Glob(filepath.Clean(dir) + "/*/*" + UpExt)
	if err != nil {
		return nil, err
	}
	files = append(files, nestedFiles...)

	var migration Migration
	for _, path := range files {
		basename := filepath.Base(path)
		dirname := filepath.Dir(path)

		downfile := path[:len(path)-len(UpExt)] + DownExt
		migration.Name = basename[:len(basename)-len(UpExt)]

		if _, err := os.Stat(downfile); os.IsNotExist(err) {
			return nil, fmt.Errorf("missing down migration: %s", downfile)
		}

		if bytes, err := ioutil.ReadFile(path); err != nil {
			return nil, fmt.Errorf("i/o: %w", err)
		} else {
			migration.UpSQL = string(bytes)
		}

		if bytes, err := ioutil.ReadFile(downfile); err != nil {
			return nil, fmt.Errorf("i/o: %w", err)
		} else {
			migration.DownSQL = string(bytes)
		}

		migration.Dir = dirname
		result = append(result, migration)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return
}
