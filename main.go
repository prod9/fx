package main

import (
	"fx.prodigy9.co/app"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/httpserver/controllers"
	"go.jonnrb.io/vanity"
)

func main() {
	handler := vanity.GitHubHandler("fx.prodigy9.co", "prod9", "fx", "https")

	err := app.Build().
		AddDefaults().
		Controllers(controllers.FromHandler("/*", handler)).
		Start()
	if err != nil {
		fxlog.Fatal(err)
	}
}
