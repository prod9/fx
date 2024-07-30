package ctrlc

import (
	"os"
	"os/signal"
	"syscall"
)

var (
	done    chan struct{}
	signals chan os.Signal
)

func Chan() <-chan struct{} {
	setup()
	return done
}

func Do(action func()) {
	setup()
	go func() {
		<-done
		action()
	}()
}

func setup() {
	if signals != nil {
		return
	}

	done = make(chan struct{})
	signals = make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-signals
		close(done)
	}()
}
