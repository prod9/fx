package prompts

import (
	"os"
	"strings"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/fxlog"
	"fx.prodigy9.co/slices"
	"github.com/mattn/go-isatty"
)

var CIConfig = config.Bool("CI")

// AlwaysYesConfig allows user to set ALWAYS_YES=1 so any confirmation prompt will result
// in a "yes" answer without requiring any user interaction. Useful in deployments, CI and
// automation scripts.
var AlwaysYesConfig = config.Bool("ALWAYS_YES")

type Session struct {
	cfg         *config.Source
	args        []string
	interactive bool
}

func New(cfg *config.Source, args []string) *Session {
	if cfg == nil {
		cfg = config.Configure()
	}

	interactive := !config.Get(cfg, CIConfig) &&
		isatty.IsTerminal(os.Stdin.Fd())
	return &Session{cfg, args, interactive}
}

func (s *Session) Len() int { return len(s.args) }

// Args return leftover unconsumed args
func (s *Session) Args() []string { return s.args }

// shift consumes and returns the next leftover arg, if any.
func (s *Session) shift() (string, bool) {
	if len(s.args) == 0 {
		return "", false
	}

	head := s.args[0]
	s.args = s.args[1:]
	return head, true
}

func (s *Session) Confirm(what, yes, no string) bool {
	if config.Get(s.cfg, AlwaysYesConfig) {
		return true
	}
	if !s.interactive {
		bailf("confirmation required: %s", what)
	}
	return readConfirm(what)
}

func (s *Session) YesNo(question string) bool {
	return s.Confirm(question, "yes", "no")
}

func (s *Session) SensitiveStr(item string) string {
	if head, ok := s.shift(); ok {
		return head
	}
	if !s.interactive {
		bailf("missing: %s", item)
	}
	return strings.TrimSpace(readSecret(item))
}

func (s *Session) OptionalStr(item string, defaultValue string) string {
	if len(s.args) > 0 {
		return s.Str(item)
	} else {
		return defaultValue
	}
}

func (s *Session) Str(item string) string {
	if head, ok := s.shift(); ok {
		return head
	}
	if !s.interactive {
		bailf("missing: %s", item)
	}
	return readLine(item)
}

func (s *Session) OptionalList(question, defaultValue string, options []string) string {
	if len(s.args) > 0 {
		return s.List(question, defaultValue, options)
	} else {
		return defaultValue
	}
}

func (s *Session) List(question, def string, options []string) string {
	if len(s.args) > 0 && slices.In(options, s.args[0]) {
		head, _ := s.shift()
		return head
	}
	if !s.interactive {
		bailf("invalid option: %s", question)
	}
	return runSelect(question, def, options)
}

// OptionalMultiSelect is MultiSelect for an optional trailing arg: with args present it
// behaves like MultiSelect, otherwise it returns defaults without prompting.
func (s *Session) OptionalMultiSelect(question string, defaults, options []string) []string {
	if len(s.args) > 0 {
		return s.MultiSelect(question, defaults, options)
	} else {
		return defaults
	}
}

// MultiSelect prompts for zero or more of options, with defaults pre-checked. From args
// it reads a single comma-separated value (each part must be a valid option); it bails
// when non-interactive with no args (mirroring List). Interactively it shows a checkbox
// menu (defaults pre-checked) toggled with space and confirmed with enter.
func (s *Session) MultiSelect(question string, defaults, options []string) []string {
	if head, ok := s.shift(); ok {
		var chosen []string
		for part := range strings.SplitSeq(head, ",") {
			if part = strings.TrimSpace(part); part == "" {
				continue
			} else if !slices.In(options, part) {
				bailf("invalid option: %s", part)
			} else {
				chosen = append(chosen, part)
			}
		}
		return chosen
	}
	if !s.interactive {
		bailf("selection required: %s", question)
	}
	return runMultiSelect(question, options, defaults)
}

func GenList[T any](s *Session, question string, def T, options []T, namer func(item T) string) T {
	names := make([]string, len(options))
	for i, item := range options {
		names[i] = namer(item)
	}

	selected := s.List(question, namer(def), names)
	for i, name := range names {
		if name == selected {
			return options[i]
		}
	}

	return def
}

func bail(err error) {
	fxlog.Fatalf("prompt: %w", err)
}

func bailf(format string, args ...any) {
	fxlog.Fatalf("prompt: "+format, args...)
}

func cancel() {
	fxlog.Fatalf("prompt: canceled")
}
