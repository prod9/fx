package cmd

import (
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"github.com/spf13/cobra"
)

func BuildServeCommand(mws []middlewares.Interface, ctrs []controllers.Interface) *cobra.Command {
	return buildServeCommand(func() *httpserver.Server {
		return httpserver.New(
			config.Configure(),
			mws,
			ctrs,
		)
	})
}

func BuildServeCommandFromFragments(fragments ...*httpserver.Fragment) *cobra.Command {
	return buildServeCommand(func() *httpserver.Server {
		return httpserver.NewWithFragments(
			config.Configure(),
			fragments,
		)
	})
}

func buildServeCommand(buildServer func() *httpserver.Server) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Starts an HTTP server.",
		RunE: func(_ *cobra.Command, args []string) error {
			return buildServer().Start()
		},
	}

}
