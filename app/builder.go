package app

import (
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"github.com/spf13/cobra"
)

type Builder struct {
	appImpl
}

func Build() *Builder           { return &Builder{} }
func (b *Builder) Start() error { return Start(&b.appImpl) }

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
func (b *Builder) DefaultAPIMiddlewares() *Builder {
	return b.Middlewares(middlewares.DefaultForAPI()...)
}
func (b *Builder) Middlewares(mws ...middlewares.Interface) *Builder {
	b.middlewares = append(b.middlewares, mws...)
	return b
}
func (b *Builder) Controllers(ctrs ...controllers.Interface) *Builder {
	b.controllers = append(b.controllers, ctrs...)
	return b
}
