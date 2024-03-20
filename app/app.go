package app

import (
	"embed"

	"fx.prodigy9.co/cmd"
	"fx.prodigy9.co/data/migrator"
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

	jobs, cmds, mws, ctrs := collect(app)
	if len(jobs) > 0 {
		cmds = append(cmds, cmd.BuildWorkerCommand(jobs...))
	}
	if len(ctrs) > 0 && len(mws) == 0 { // auto-default some middlewares if controllers are added
		mws = middlewares.DefaultForAPI()
	}
	if len(ctrs) > 0 || len(mws) > 0 { // don't add `serve` command if there's nothing to serve
		cmds = append(cmds, cmd.BuildServeCommand(mws, ctrs))
	}

	return cmd.
		BuildRootCommand(app.Description(), cmds...).
		Execute()
}

func collect(app Interface) (
	[]worker.Interface,
	[]*cobra.Command,
	[]middlewares.Interface,
	[]controllers.Interface,
) {
	var (
		jobs = app.Jobs()
		mws  = app.Middlewares()
		ctrs = app.Controllers()
		cmds = app.Commands()
	)

	for _, child := range app.Children() {
		childJobs, childCmds, childMws, childCtrs := collect(child)
		jobs = append(jobs, childJobs...)
		cmds = append(cmds, childCmds...)
		mws = append(mws, childMws...)
		ctrs = append(ctrs, childCtrs...)
	}

	return jobs, cmds, mws, ctrs
}
