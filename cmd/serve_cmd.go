package cmd

import (
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/httpserver"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
	"github.com/spf13/cobra"
)

func BuildServeCommand(mws []middlewares.Interface, ctrs []controllers.Interface) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Starts an HTTP server.",
		RunE: func(_ *cobra.Command, args []string) error {
			cfg := config.Configure()
			srv := httpserver.New(cfg, mws, ctrs)
			return srv.Start()
		},
	}
}
