package main

import (
	"fx.prodigy9.co/app"
	"fx.prodigy9.co/cmd"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/pubsub"
)

// Ball is the payload volleyed between the two processes; Rally counts the exchanges.
type Ball struct {
	Rally int
}

// Ping carries the ball toward the ponger, Pong carries it back. The two processes share
// no memory and no socket — the whole rally runs over Postgres LISTEN/NOTIFY, one
// subscription each. Names follow their var, the pubsub convention.
var (
	Ping = pubsub.NewChannel[Ball]("pingpong_ping")
	Pong = pubsub.NewChannel[Ball]("pingpong_pong")
)

// Run from this directory (so its .env, and its own database, are picked up). Create the
// database once, then start the two processes in either order:
//
//	cd examples/pingpong
//	go run . data create-db
//	go run . ponger   # in one shell
//	go run . pinger   # in another
//
// The pinger serves a ball and turns every pong into the next ping; the ponger echoes
// every ping back.
func main() {
	err := app.Build().
		Name("pingpong").
		Command(PingerCmd).
		Command(PongerCmd).
		Command(cmd.BuildDataCommand()).
		Command(cmd.PrintConfigCmd).
		Start()
	if err != nil {
		fxlog.Fatal(err)
	}
}
