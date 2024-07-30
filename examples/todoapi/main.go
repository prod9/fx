package main

import (
	"log"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/examples/todoapi/auth"
	"fx.prodigy9.co/examples/todoapi/todos"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/httpserver/middlewares"
)

func main() {
	err := app.Build().
		Description("Example TODO API application").
		DefaultAPIMiddlewares().
		Middlewares(middlewares.DebugRequest).
		Controllers(controllers.Home{}).
		Mount(auth.App).
		Mount(todos.App).
		Start()

	if err != nil {
		log.Fatalln(err)
	}
}
