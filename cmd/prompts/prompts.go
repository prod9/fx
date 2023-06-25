package prompts

import (
	"log"
	"strings"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/slices"
	"github.com/pterm/pterm"
)

var CIConfig = config.Bool("CI")

// AlwaysYesConfig allows user to set ALWAYS_YES=1 so any confirmation prompt will result
// in a "yes" answer without requiring any user interaction. Useful in deployments, CI and
// automation scripts.
var AlwaysYesConfig = config.Bool("ALWAYS_YES")

type Session struct {
	cfg  *config.Source
	args []string
}

func New(cfg *config.Source, args []string) *Session {
	if cfg == nil {
		cfg = config.Configure()
	}
	return &Session{cfg, args}
}

func (s *Session) Len() int {
	return len(s.args)
}

// Args return leftover unconsumed args
func (s *Session) Args() []string {
	return s.args
}

func (s *Session) Confirm(what, yes, no string) bool {
	if config.Get(s.cfg, CIConfig) {
		return true
	}
	if config.Get(s.cfg, AlwaysYesConfig) {
		return true
	}

	result, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText(what).
		WithConfirmText(yes).
		WithRejectText(no).
		Show()
	if err != nil {
		log.Fatalln(err)
		return false
	} else {
		return result
	}
}

func (s *Session) YesNo(question string) bool {
	return s.Confirm(question, "yes", "no")
}

func (s *Session) SensitiveStr(item string) string {
	if len(s.args) > 0 {
		head, tail := s.args[0], s.args[1:]
		s.args = tail
		return head
	}

	result, err := pterm.DefaultInteractiveTextInput.
		WithMask("*").
		WithDefaultText(item).
		Show()
	if err != nil {
		log.Fatalln(err)
		return ""
	} else {
		return strings.TrimSpace(result)
	}
}

func (s *Session) OptionalStr(item string, defaultValue string) string {
	if len(s.args) > 0 {
		return s.Str(item)
	} else {
		return defaultValue
	}
}

func (s *Session) Str(item string) string {
	if len(s.args) > 0 {
		head, tail := s.args[0], s.args[1:]
		s.args = tail
		return head
	}

	result, err := pterm.DefaultInteractiveTextInput.
		WithDefaultText(item).
		Show()
	if err != nil {
		log.Fatalln(err)
		return ""
	} else {
		return strings.TrimSpace(result)
	}
}

func (s *Session) List(question, def string, options []string) string {
	if len(s.args) > 0 {
		head, tail := s.args[0], s.args[1:]
		if slices.In(options, head) {
			s.args = tail
			return head
		}
	}

	result, err := pterm.DefaultInteractiveSelect.
		WithDefaultText(question).
		WithOptions(options).
		WithDefaultOption(def).
		Show()
	if err != nil {
		log.Fatalln(err)
		return ""
	} else {
		return result
	}
}
