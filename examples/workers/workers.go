package main

import (
	"errors"

	"fx.prodigy9.co/app"
	"fx.prodigy9.co/cmd"
	datacmd "fx.prodigy9.co/cmd/data"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/worker"
)

func main() {
	err := app.Build().
		Job(&Reporter{}).
		Job(&Creator{}).
		Job(&Incrementer{}).
		Command(SpawnCmd).
		Command(datacmd.Cmd).
		Command(cmd.PrintConfigCmd).
		Start()
	if err != nil {
		if errors.Is(err, worker.ErrStop) {
			fxlog.Log("stopped")
		} else {
			fxlog.Fatal(err)
		}
	}
}
