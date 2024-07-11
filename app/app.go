package app

import (
	"embed"

	"fx.prodigy9.co/cmd"
	"fx.prodigy9.co/data/migrator"
	"fx.prodigy9.co/httpserver"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
)

type Interface interface {
	Name() string
	Description() string
	Children() []Interface

	Commands() []*cobra.Command
	EmbeddedMigrations() *embed.FS
	Jobs() []worker.Interface

	Middlewares() []middlewares.Interface
	Controllers() []controllers.Interface
}

func Start(app Interface) error {
	if app.EmbeddedMigrations() != nil {
		migrator.Embed(*app.EmbeddedMigrations())
	}

	jobs, cmds, fragment := collect(app)
	if len(jobs) > 0 {
		cmds = append(cmds, cmd.BuildWorkerCommand(jobs...))
	}
	if !fragment.IsEmpty() {
		if fragment.HasNoMiddlewares() {
			fragment.AddMiddlewares(middlewares.DefaultForAPI()...)
		}
		cmds = append(cmds, cmd.BuildServeCommandFromFragments(fragment))
	}

	return cmd.
		BuildRootCommand(app.Description(), cmds...).
		Execute()
}

func collect(app Interface) (
	[]worker.Interface,
	[]*cobra.Command,
	*httpserver.Fragment,
) {
	var (
		jobs     = app.Jobs()
		cmds     = app.Commands()
		fragment = httpserver.NewFragment(
			app.Middlewares(),
			app.Controllers(),
		)
	)

	for _, child := range app.Children() {
		childJobs, childCmds, childFragment := collect(child)
		jobs = append(jobs, childJobs...)
		cmds = append(cmds, childCmds...)
		fragment.AddChild(childFragment)
	}

	return jobs, cmds, fragment
}
