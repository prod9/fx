package main

import (
	"log"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/examples/todoapi/auth"
	"fx.prodigy9.co/httpserver/controllers"
)

func main() {
	err := app.Build().
		Description("Example TODO API application").
		DefaultAPIMiddlewares().
		Controllers(controllers.Home{}).
		Mount(auth.App).
		Start()

	if err != nil {
		log.Fatalln(err)
	}
}
