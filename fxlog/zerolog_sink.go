package fxlog

import (
	"log/slog"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type ZerologSink struct {
	zerolog zerolog.Logger
}

var _ Sink = (*ZerologSink)(nil)

func NewZerologSink(zerolog zerolog.Logger) *ZerologSink {
	return &ZerologSink{
		zerolog: zerolog,
	}
}

func NewDefaultZerologSink() *ZerologSink {
	zl := zerolog.
		New(zerolog.ConsoleWriter{
			TimeFormat: time.RFC3339Nano,
			Out:        os.Stderr,
		}).
		With().Timestamp().
		Logger()
	return NewZerologSink(zl)
}

func (s *ZerologSink) Log(msg string, attrs ...slog.Attr) {
	if len(attrs) == 0 {
		s.zerolog.Info().Msg(msg)
		return
	}

	ev := s.zerolog.Info()
	for _, attr := range attrs {
		switch attr.Value.Kind() {
		case slog.KindBool:
			ev = ev.Bool(attr.Key, attr.Value.Bool())
		case slog.KindDuration:
			ev = ev.Dur(attr.Key, attr.Value.Duration())
		case slog.KindFloat64:
			ev = ev.Float64(attr.Key, attr.Value.Float64())
		case slog.KindInt64:
			ev = ev.Int64(attr.Key, attr.Value.Int64())
		case slog.KindString:
			ev = ev.Str(attr.Key, attr.Value.String())
		case slog.KindTime:
			ev = ev.Time(attr.Key, attr.Value.Time())
		case slog.KindUint64:
			ev = ev.Uint64(attr.Key, attr.Value.Uint64())
		case slog.KindGroup:
			bail("slog.KindGroup is not supported")
		case slog.KindLogValuer:
			bail("slog.KindLogValuer is not supported")
		default:
			bail("slog kind %s is not supported", attr.Value.Kind())
		}
	}

	ev.Msg(msg)
}

func (s *ZerologSink) Error(err error) {
	s.zerolog.Error().Err(err).Msg("error")
}

func (s *ZerologSink) Fatal(err error) {
	s.zerolog.Fatal().Err(err).Msg("fatal")
}
