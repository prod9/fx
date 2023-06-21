package migrator

import (
	"context"
	"fmt"

	"fx.prodigy9.co/data"
	"github.com/chakrit/gendiff"
	"github.com/jmoiron/sqlx"
)

// language=PostgreSQL
const (
	CreateMigrationsTableSQL = `
		CREATE TABLE IF NOT EXISTS migrations
		(
			name     text PRIMARY KEY,
			up_sql   text NOT NULL,
			down_sql text NOT NULL
		);`
	ListMigrationsSQL = `
		SELECT *
		FROM migrations
		ORDER BY name ASC`
	UpdateMigrationSQL = `
		INSERT INTO migrations (name, up_sql, down_sql)
		VALUES ($1, $2, $3)
		ON CONFLICT (name) DO UPDATE
		SET up_sql = $2, down_sql = $3`
	PruneMigrationSQL = `
		DELETE FROM migrations
		WHERE name = $1`
)

const (
	MaxRollbacks = 1
	UpExt        = ".up.sql"
	DownExt      = ".down.sql"
)

type Migrator struct {
	db  *sqlx.DB
	dir string
}

func New(db *sqlx.DB, dir string) *Migrator {
	return &Migrator{db, dir}
}

func (m *Migrator) Plan(ctx context.Context, intent Intent) (actions []Plan, dirty bool, err error) {
	var (
		scope   data.Scope
		inFiles []Migration
		inDB    []Migration
	)

	if scope, err = data.NewScope(ctx, m.db); err != nil {
		return
	} else {
		defer scope.End(&err)
	}

	if inFiles, err = LoadMigrations(m.dir); err != nil {
		return
	} else if inDB, err = m.loadFromDB(scope.Context()); err != nil {
		return
	}

	switch intent {
	case IntentSync:
		return m.planSync(inFiles, inDB)
	case IntentMigrate:
		return m.planMigrate(inFiles, inDB)
	case IntentRollback:
		return m.planRollback(inFiles, inDB)
	default:
		return nil, false, nil
	}
}

func (m *Migrator) planSync(inFiles []Migration, inDB []Migration) (actions []Plan, dirty bool, err error) {
	diffs := gendiff.Make(migrationDiff{inDB, inFiles})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Delete:
			dirty = true
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				actions = append(actions, Plan{ActionPrune, inDB[lidx]})
			}

		case gendiff.Match:
			lidx, ridx := d.Lstart, d.Rstart
			for lidx < d.Lend && ridx < d.Rend {
				mDB, mFile := inDB[lidx], inFiles[ridx]
				if mDB.UpSQL != mFile.UpSQL || mDB.DownSQL != mFile.DownSQL {
					dirty = true
					actions = append(actions, Plan{ActionUpdate, mFile})
				}
				lidx += 1
				ridx += 1
			}
		}
	}

	return
}

func (m *Migrator) planMigrate(inFiles []Migration, inDB []Migration) (actions []Plan, dirty bool, err error) {
	diffs := gendiff.Make(migrationDiff{inDB, inFiles})

	for _, d := range diffs {
		switch d.Op {
		case gendiff.Insert:
			// some migrations were removed/changed prior to this migration, which means that
			// the db is likely not in the state that the migration expects it to be.
			if dirty {
				err = fmt.Errorf("db state divergence detected, please carefully review and re-sync")
				return
			}
			for ridx := d.Rstart; ridx < d.Rend; ridx++ {
				actions = append(actions, Plan{ActionMigrate, inFiles[ridx]})
			}

		case gendiff.Delete:
			dirty = true
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				actions = append(actions, Plan{ActionPrune, inDB[lidx]})
			}

		case gendiff.Match:
			lidx, ridx := d.Lstart, d.Rstart
			for lidx < d.Lend && ridx < d.Rend {
				mDB, mFile := inDB[lidx], inFiles[ridx]
				if mDB.UpSQL != mFile.UpSQL || mDB.DownSQL != mFile.DownSQL {
					dirty = true
					actions = append(actions, Plan{ActionUpdate, mFile})
				}
				lidx += 1
				ridx += 1
			}
		}
	}

	return
}

func (m *Migrator) planRollback(inFiles []Migration, inDB []Migration) (actions []Plan, dirty bool, err error) {
	rollbackIdx := len(inDB) - MaxRollbacks

	diffs := gendiff.Make(migrationDiff{inDB, inFiles})
	for _, d := range diffs {
		switch d.Op {
		case gendiff.Insert:
			if d.Lstart <= rollbackIdx {
				err = fmt.Errorf("db state divergence detected, please carefully review and re-sync")
				return
			}

		case gendiff.Delete:
			dirty = true
			if d.Lstart <= rollbackIdx {
				err = fmt.Errorf("db state divergence detected, please carefully review and re-sync")
				return
			}
			for lidx := d.Lstart; lidx < d.Lend; lidx++ {
				actions = append(actions, Plan{ActionPrune, inDB[lidx]})
			}

		case gendiff.Match:
			lidx, ridx := d.Lstart, d.Rstart
			for lidx < d.Lend && ridx < d.Rend {
				mDB, mFile := inDB[lidx], inFiles[ridx]
				if mDB.UpSQL != mFile.UpSQL || mDB.DownSQL != mFile.DownSQL {
					dirty = true
					if d.Lstart <= rollbackIdx {
						err = fmt.Errorf("db state divergence detected, please carefully review and re-sync")
						return
					}
					actions = append(actions, Plan{ActionUpdate, mFile})
				}
				lidx += 1
				ridx += 1
			}
		}
	}

	// all prior states to the rollbacks seems OK, performing rollbacks (in reverse)
	for idx := len(inDB) - 1; idx >= rollbackIdx; idx-- {
		actions = append(actions, Plan{ActionRollback, inDB[idx]})
	}
	return
}

func (m *Migrator) Apply(ctx context.Context, plan Plan) (err error) {
	var (
		scope data.Scope
		mig   Migration
	)
	if scope, err = data.NewScope(ctx, m.db); err != nil {
		return
	} else {
		defer scope.End(&err)
	}

	mig = plan.Migration

	switch plan.Action {
	case ActionUpdate:
		if err = scope.Exec(UpdateMigrationSQL, mig.Name, mig.UpSQL, mig.DownSQL); err != nil {
			return
		}

	case ActionIgnore:
		// no-op

	case ActionPrune:
		if err = scope.Exec(PruneMigrationSQL, mig.Name); err != nil {
			return
		}

	case ActionMigrate:
		if err = scope.Exec(UpdateMigrationSQL, mig.Name, mig.UpSQL, mig.DownSQL); err != nil {
			return
		} else if err = scope.Exec(plan.Migration.UpSQL); err != nil {
			return
		}

	case ActionRollback:
		if err = scope.Exec(plan.Migration.DownSQL); err != nil {
			return
		} else if err = scope.Exec(PruneMigrationSQL, mig.Name); err != nil {
			return
		}

	default:
		err = fmt.Errorf("unknown Action in plan: %d", plan.Action)
	}

	return
}

func (m *Migrator) loadFromDB(ctx context.Context) (result []Migration, err error) {
	var scope data.Scope
	if scope, err = data.NewScope(ctx, m.db); err != nil {
		return nil, err
	} else {
		defer scope.End(&err)
	}

	if err = scope.Exec(CreateMigrationsTableSQL); err != nil {
		return
	} else if err = scope.Select(&result, ListMigrationsSQL); err != nil {
		return
	}

	return
}
