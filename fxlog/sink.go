package fxlog

import (
	"fx.prodigy9.co/config"
	"log/slog"
)

// Sink as a logging message sink creates a simple abstraction to allow swapping logging
// implementations to be swappable in and out without requiring changes to fx itself.
//
// The logging convention in fx follows Go's standard library's convention in the sense
// of not doing the complex log-level juggling, instead focusing on making log outputs
// as simple and as useful as possible rather than requiring devs or operators to have to
// do a log of configuration to get useful output.
type Sink interface {
	Log(msg string, attrs ...slog.Attr)
	Error(err error)
	Fatal(err error)
}

var (
	LogSinkConfig      = config.StrDef("LOG_SINK", "zerolog")
	activeSink    Sink = nil
)

func SetSink(sink Sink) {
	activeSink = sink
}

func sink() Sink {
	if activeSink != nil {
		return activeSink
	}

	cfg := config.Configure()

	sinkConfig := config.Get(cfg, LogSinkConfig)
	switch sinkConfig {
	case "slog":
		activeSink = NewDefaultSLogSink()
	case "zerolog":
		activeSink = NewDefaultZerologSink()
	default:
		bail("invalid sink %s", sinkConfig)
		activeSink = NewDefaultSLogSink() // default to slog
	}

	return activeSink
}
