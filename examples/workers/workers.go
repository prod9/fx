package main

import (
	"log"

	"fx.prodigy9.co/app"
)

func main() {
	err := app.Build().
		Job(&Reporter{}).
		Job(&Creator{}).
		Job(&Incrementer{}).
		Command(SpawnCmd).
		Start()
	if err != nil {
		log.Fatalln(err)
	}
}
