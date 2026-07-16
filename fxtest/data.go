package fxtest

import (
	"context"
	"strings"
	"testing"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/data/dbname"
)

// FXTEST_CLEANUP can be set to `none` to stop test databases from being dropped so
// they can be inspected after test runs for debugging purposes.
//
// Afterwards, it can be set to `force` to force dropping of the test databases before
// creating new ones.
var CleanupConfig = config.Str("FXTEST_CLEANUP")

// FXTEST_SKIP_DBTESTS can be set to `true` (or `1`) to skip every test that requires a
// database, for hermetic environments with no database such as in-build test gates.
var SkipDBTestsConfig = config.Bool("FXTEST_SKIP_DBTESTS")

func ConnectTestDatabase(t *testing.T) context.Context {
	cfg := Configure()
	if config.Get(cfg, SkipDBTestsConfig) {
		t.Skip("fxtest: FXTEST_SKIP_DBTESTS is set")
		return nil
	}
	cleanupMode := strings.ToUpper(config.Get(cfg, CleanupConfig))

	dbURL := config.Get(cfg, data.DatabaseURLConfig)
	if dbURL == "" {
		t.Log("fxtest: DATABASE_URL is required to run tests")
		t.FailNow()
		return nil
	}

	name, err := dbname.From(dbURL)
	if err != nil {
		t.Logf("fxtest: %s", err)
		t.FailNow()
		return nil
	}

	name += "_" + t.Name()
	name = dbname.Sanitize(name)
	dbURL, err = dbname.Set(dbURL, name)
	if err != nil {
		t.Logf("fxtest: %s", err)
		t.FailNow()
		return nil
	}

	config.Set(cfg, data.DatabaseURLConfig, dbURL)

	if cleanupMode == "FORCE" {
		t.Log("fxtest: dropping database", name)
		if err := data.DropDB(cfg); err != nil {
			if !strings.Contains(err.Error(), "does not exist") {
				t.Logf("fxtest: %s", err)
				t.FailNow()
				return nil
			}
		}
	}

	t.Log("fxtest: creating database", name)
	if err := data.CreateDB(cfg); err != nil {
		t.Logf("fxtest: %s", err)
		t.FailNow()
		return nil
	}
	if cleanupMode != "NONE" {
		t.Cleanup(func() {
			t.Log("fxtest: dropping database", name)
			if err := data.DropDB(cfg); err != nil {
				t.Logf("fxtest: %s", err)
				t.FailNow()
			}
		})
	}

	db, err := data.Connect(cfg)
	if err != nil {
		t.Logf("fxtest: %s", err)
		t.FailNow()
		return nil
	}
	t.Cleanup(func() {
		err := db.Close()
		if err != nil {
			t.Logf("fxtest: %s", err)
			t.FailNow()
		}
	})

	ctx := t.Context()
	ctx = config.NewContext(ctx, cfg)
	ctx = data.NewContext(ctx, db)
	return ctx
}
