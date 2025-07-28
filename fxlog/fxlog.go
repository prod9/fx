// fxlog provides a centralized logging sink for everything in the fx package in a way
// that matches the context usage convention in other parts of the codebase.
//
// If an active sink is not set, it will default to using zerolog as the sink to get
// pretty console output in a performant way.
//
// A different sink can be set by calling `fxlog.SetSink` with a fxlog.Sink
// implementation. Default implementation for log/slog and zerolog is provided via
// NewSlogSink and NewZerologSink respectively.
//
// For package authors, use the Log, Error, and Fatal functions to log messages.
package fxlog

import (
	"fmt"
	"log/slog"
	"os"
)

// Import slog.* attribute functions and other useful methods usually needed to build
// log messages. This allows codebase to focus on just using the `fxlog` package without
// having to know which other package needs to be imported.
var (
	Any      = slog.Any
	String   = slog.String
	Int64    = slog.Int64
	Int      = slog.Int
	Uint64   = slog.Uint64
	Float64  = slog.Float64
	Bool     = slog.Bool
	Time     = slog.Time
	Duration = slog.Duration
)

func Log(msg string, attrs ...slog.Attr) { sink().Log(msg, attrs...) }
func Error(err error)                    { sink().Error(err) }
func Errorf(msg string, args ...any)     { Error(fmt.Errorf(msg, args...)) }
func Fatal(err error)                    { sink().Fatal(err) }
func Fatalf(msg string, args ...any)     { Fatal(fmt.Errorf(msg, args...)) }

func bail(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, "fxlog: "+msg, args...)
}
