package app

import (
	"embed"

	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"github.com/spf13/cobra"
)

type appImpl struct {
	name          string
	description   string
	configuration map[string]any
	children      []Interface

	rootCommand *cobra.Command
	commands    []*cobra.Command
	migrations  *embed.FS

	middlewares []middlewares.Interface
	controllers []controllers.Interface
}

var _ Interface = &appImpl{}

func (a *appImpl) Name() string                   { return a.name }
func (a *appImpl) Description() string            { return a.description }
func (a *appImpl) Configurations() map[string]any { return a.configuration }
func (a *appImpl) Children() []Interface          { return a.children }

func (a *appImpl) RootCommand() *cobra.Command   { return a.rootCommand }
func (a *appImpl) Commands() []*cobra.Command    { return a.commands }
func (a *appImpl) EmbeddedMigrations() *embed.FS { return a.migrations }

func (a *appImpl) Middlewares() []middlewares.Interface { return a.middlewares }
func (a *appImpl) Controllers() []controllers.Interface { return a.controllers }
