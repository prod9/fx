package main

import (
	"log"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/app/files"
	"fx.prodigy9.co/examples/sendfile/drops"
	"fx.prodigy9.co/httpserver/controllers"
)

func main() {
	err := app.Build().
		Description("Send a file to a friend — files + blobserver example").
		AddDefaults().
		Controllers(controllers.Home{}).
		Mount(files.App).
		Mount(drops.App).
		Start()

	if err != nil {
		log.Fatalln(err)
	}
}
