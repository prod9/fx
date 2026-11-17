package app

import (
	"context"
	"embed"

	"fx.prodigy9.co/buildinfo"
	"fx.prodigy9.co/cmd"
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

// Start is the reference assembly over the collectors: any caller can run the same
// collectors against its own root command instead.
func Start(app Interface) error {
	RegisterMigrations(app)

	cmds := CollectCommands(app)
	if jobs := CollectJobs(app); len(jobs) > 0 {
		cmds = append(cmds, cmd.BuildWorkerCommand(jobs...))
	}
	if fragment := CollectFragment(app); !fragment.IsEmpty() {
		if fragment.HasNoMiddlewares() {
			fragment.AddMiddlewares(middlewares.DefaultForAPI()...)
		}
		cmds = append(cmds, cmd.BuildServeCommandFromFragments(fragment))
	}

	info := buildinfo.Get()
	if name := app.Name(); name != "" {
		info = info.WithName(name)
	}
	ctx := buildinfo.NewContext(context.Background(), info)

	return cmd.
		BuildRootCommand(app.Description(), cmds...).
		ExecuteContext(ctx)
}
