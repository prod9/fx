package migrator

import (
	"fmt"
	"strings"
)

type Action int

const (
	// ActionUpdate updates the migration content in the database
	ActionUpdate = Action(iota)
	// ActionIgnore ignores a new migration file that has yet to be run.
	ActionIgnore
	// ActionPrune prunes missing migration entries from the database
	ActionPrune
	// Recover the migration content from the database to a file
	ActionRecover
	// ActionMigrate runs all migrations that are yet to be ran
	ActionMigrate
	// ActionRollback rollbacks the most recent migration
	ActionRollback
)

func (act Action) String() string {
	switch act {
	case ActionUpdate:
		return "update sql"
	case ActionIgnore:
		return "ignore"
	case ActionPrune:
		return "remove"
	case ActionRecover:
		return "recover"
	case ActionMigrate:
		return "migrate"
	case ActionRollback:
		return "rollback"
	default:
		return "(unknown)"
	}
}

type Intent int

const (
	// TODO: IntentRecover ?
	IntentSync = Intent(iota)
	IntentMigrate
	IntentRollback
)

type Plan struct {
	Action    Action
	Migration Migration
}

func (p Plan) String() string {
	date, name := "", ""
	parts := strings.SplitN(p.Migration.Name, "_", 2)
	switch len(parts) {
	case 0:
		date, name = "209901010000", "(unknown)"
	case 1:
		date, name = parts[0], "(unknown)"
	default:
		date, name = parts[0], parts[1]
	}

	return fmt.Sprintf("%20s => %s %s", p.Action, date, name)
}
