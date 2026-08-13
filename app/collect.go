package app

import (
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/httpserver"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
)

// The collectors walk a finished app tree and hand back plain values, so a caller can
// compose commands onto any root — Start is one reference consumer, not the only door.

func CollectCommands(app Interface) []*cobra.Command {
	cmds := app.Commands()
	for _, child := range app.Children() {
		cmds = append(cmds, CollectCommands(child)...)
	}
	return cmds
}

func CollectJobs(app Interface) []worker.Interface {
	jobs := app.Jobs()
	for _, child := range app.Children() {
		jobs = append(jobs, CollectJobs(child)...)
	}
	return jobs
}

func CollectFragment(app Interface) *httpserver.Fragment {
	fragment := httpserver.NewFragment(app.Middlewares(), app.Controllers())
	for _, child := range app.Children() {
		fragment.AddChild(CollectFragment(child))
	}
	return fragment
}

// RegisterMigrations registers every fragment's embedded migrations into the migrator's
// global registry — the sanctioned channel LoadAuto and the data command read at run
// time. Start calls it; callers composing without Start call it themselves.
func RegisterMigrations(app Interface) {
	if mig := app.EmbeddedMigrations(); mig != nil {
		migrator.Embed(*mig)
	}
	for _, child := range app.Children() {
		RegisterMigrations(child)
	}
}
