package main

import (
	"context"
	"log"
	"math/rand"
	"strconv"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"fx.prodigy9.co/worker"
	"github.com/spf13/cobra"
)

var SpawnCmd = &cobra.Command{
	Use:   "spawn-jobs",
	Short: "Spawn initial set of jobs",
	Run:   runSpawnCmd,
}

func runSpawnCmd(cmd *cobra.Command, args []string) {
	cfg := config.Configure()
	db, err := data.Connect(cfg)
	if err != nil {
		log.Println(err)
	}

	// clear existing jobs since the job themselves are self-rescheduling there are bound
	// to be existing tasks in the queue
	if _, err := db.Exec("TRUNCATE jobs"); err != nil {
		log.Fatalln(err)
	}

	ctx := context.Background()
	ctx = data.NewContext(ctx, db)
	jobs := []worker.Interface{
		&Creator{Count: 20},
		&Reporter{},
	}

	for i := 1; i <= 20; i++ {
		jobs = append(jobs, &Incrementer{
			Counter:     "counter" + strconv.Itoa(i),
			IncrementBy: rand.Intn(20) + 1,
		})
	}

	for _, job := range jobs {
		if _, err := worker.ScheduleNow(ctx, job); err != nil {
			log.Fatalln(err)
		}
	}
}
