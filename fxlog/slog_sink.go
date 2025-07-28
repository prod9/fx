package fxlog

import (
	"log/slog"
	"os"
)

type SLogSink struct {
	slog *slog.Logger
}

var _ Sink = (*SLogSink)(nil)

func NewSLogSink(slog *slog.Logger) *SLogSink {
	return &SLogSink{
		slog: slog,
	}
}

func NewDefaultSLogSink() *SLogSink {
	return NewSLogSink(slog.Default())
}

func (s *SLogSink) Log(msg string, attrs ...slog.Attr) {
	s.slog.LogAttrs(nil, slog.LevelInfo, msg, attrs...)
}

func (s *SLogSink) Error(err error) {
	s.slog.LogAttrs(nil, slog.LevelError, "error", slog.Any("error", err))
}

func (s *SLogSink) Fatal(err error) {
	s.slog.LogAttrs(nil, slog.LevelError, "fatal", slog.Any("error", err))

	// TODO: Ensure the log message is flushed before the exit
	os.Exit(1)
}
