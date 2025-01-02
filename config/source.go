package config

import (
	"log"
)

const DotEnvSearchLimit = 4

var defaultSource *Source = &Source{}

func DefaultSource() *Source {
	return defaultSource
}

// Source represents a source of configuration values. It is a combined collection of
// predefined configuration variables along with the Provider to provider their values.
//
// Source can be used with the pakage-level getter/setter functions to obtain
// actual configuration values.
//
// Source (alongs with config.XXX builders) allow parts of the application (and fx itself)
// to provide its own set of configuration variables without requiring direct changes to
// the `config` package everytime a new configuration is needed.
type Source struct {
	provider Provider
	vars     []_Var
}

// NewSource create a new, empty Source.
func NewSource(provider Provider, vars []_Var) *Source {
	if provider == nil {
		provider = EnvProvider{}
	}
	if err := provider.Initialize(); err != nil {
		log.Println("config:", err)
		// continue, don't block application start
	}
	return &Source{provider, vars}
}

// Configure creates a Source that automatically includes all `Var`iables declared in the
// application through NewVar or the typed variants along with the default provider.
//
// Variables can be created with various typed functions in the package like `.Str` or
// `.Bool`.
//
// Default provider can be overridden with the `SetDefaultProvider` method. Alternatively,
// it is not difficult to roll your own `Configure()` with custom provider and variable
// sets using the `NewSource()` method.
func Configure() *Source {
	return NewSource(
		defaultSource.provider,
		defaultSource.vars,
	)
}

func (s *Source) Provider() Provider {
	return s.provider
}
func (s *Source) Vars() []_Var {
	return s.vars
}
