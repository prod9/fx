package main

import (
	"context"
	"log"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/app/files"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/examples/sendfile/drops"
	"fx.prodigy9.co/httpserver/controllers"
	"fx.prodigy9.co/worker"
)

func main() {
	go ensureBackgroundJobs(files.CleanupJob)

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

func ensureBackgroundJobs(jobs ...worker.Interface) {
	cfg := config.Configure()
	db, err := data.Connect(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := data.NewContext(context.Background(), db)
	for _, job := range jobs {
		if _, err := worker.ScheduleNowIfNotExists(ctx, job); err != nil {
			log.Fatalln(err)
		}
	}
}
