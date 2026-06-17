package fxlog

import (
	"context"
	"log/slog"
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
	s.slog.LogAttrs(context.TODO(), slog.LevelInfo, msg, attrs...)
}

func (s *SLogSink) Error(err error) {
	s.slog.LogAttrs(context.TODO(), slog.LevelError, "error", slog.Any("error", err))
}
