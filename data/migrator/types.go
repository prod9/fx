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
	// Exports the migration content from the database to a file
	// TODO: ActionExport
	//   This would be a different mode for `ActionPrune` so we'll need a switch
	//   for user to specify which action they want for migrations that is in
	//   the database but is missing from the filesystem. IMO This is a better
	//   default option than simply pruning as well. Also, if the switch is on,
	//   ActionUpdate should result in updating the filesystem from the DB rather
	//   than the other way around.

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
