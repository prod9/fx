package ctrlc

import (
	"os"
	"os/signal"
	"syscall"
)

func Do(action func()) {
	ctrlC := make(chan os.Signal, 1)
	signal.Notify(ctrlC, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-ctrlC
		action()
	}()
}
