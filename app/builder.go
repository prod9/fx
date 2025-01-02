package app

import (
	"embed"

	"fx.prodigy9.co/cmd"
	"fx.prodigy9.co/cmd/data"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
)

type Builder struct {
	appImpl
}

func Build() *Builder           { return &Builder{} }
func (b *Builder) Start() error { return Start(&b.appImpl) }

func (b *Builder) AddDefaults() *Builder {
	return b.AddDefaultMiddlewares().
		AddDefaultCommands()
}
func (b *Builder) AddDefaultMiddlewares() *Builder {
	return b.Middlewares(middlewares.DefaultForAPI()...)
}
func (b *Builder) AddDefaultCommands() *Builder {
	return b.Commands(
		cmd.PrintConfigCmd,
		cmd.TestEmailCmd,
		data.Cmd,
	)
}

func (b *Builder) Name(name string) *Builder {
	b.name = name
	return b
}
func (b *Builder) Description(description string) *Builder {
	b.description = description
	return b
}
func (b *Builder) Mount(builder *Builder) *Builder {
	b.children = append(b.children, &builder.appImpl)
	return b
}

func (b *Builder) Command(cmd *cobra.Command) *Builder {
	b.commands = append(b.commands, cmd)
	return b
}
func (b *Builder) Commands(cmds ...*cobra.Command) *Builder {
	b.commands = append(b.commands, cmds...)
	return b
}
func (b *Builder) EmbedMigrations(migrations embed.FS) *Builder {
	b.migrations = &migrations
	return b
}
func (b *Builder) Job(job worker.Interface) *Builder {
	b.jobs = append(b.jobs, job)
	return b
}

func (b *Builder) Middlewares(mws ...middlewares.Interface) *Builder {
	b.middlewares = append(b.middlewares, mws...)
	return b
}
func (b *Builder) Controllers(ctrs ...controllers.Interface) *Builder {
	b.controllers = append(b.controllers, ctrs...)
	return b
}
