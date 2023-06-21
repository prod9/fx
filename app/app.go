package app

import (
	"fx.prodigy9.co/cmd"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"github.com/spf13/cobra"
)

type Interface interface {
	Name() string
	Description() string
	Children() []Interface

	Commands() []*cobra.Command
	Middlewares() []middlewares.Interface
	Controllers() []controllers.Interface
}

func Start(app Interface) error {
	cmds, mws, ctrs := collect(app)
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

func collect(app Interface) ([]*cobra.Command, []middlewares.Interface, []controllers.Interface) {
	var (
		mws  = app.Middlewares()
		ctrs = app.Controllers()
		cmds = app.Commands()
	)

	for _, child := range app.Children() {
		childCmds, childMws, childCtrs := collect(child)
		cmds = append(cmds, childCmds...)
		mws = append(mws, childMws...)
		ctrs = append(ctrs, childCtrs...)
	}

	return cmds, mws, ctrs
}
