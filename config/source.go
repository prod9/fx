package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

var defaultSource *Source = NewSource(nil)

// Source represents a source of configuration values. You can Get configuration values
// from it. At the moment it simply gets value from os.Env ensuring that godotenv is
// loaded before doing so. In the future it could be changed or overloaded to provide
// values from etcd or hashicorp vault, for example. The getter/setter of values are
// actually implemented as package functions, for now, due to limitations on Go generics.
//
// Source (alongs with config.XXX builders) allow parts of the application (and fx itself)
// to provide its own set of configuration variables without requiring direct changes to
// the `config` package everytime a new configuration is needed.
type Source struct {
	vars []_Var
}

// NewSource create a new, empty Source.
func NewSource(vars []_Var) *Source {
	var envs []string
	if _, err := os.Stat(".env.local"); !os.IsNotExist(err) {
		envs = append(envs, ".env.local")
	}
	if _, err := os.Stat(".env"); !os.IsNotExist(err) {
		envs = append(envs, ".env")
	}

	if len(envs) > 0 {
		if err := godotenv.Load(envs...); err != nil {
			log.Println("dotenv:", err)
			return &Source{vars}
		}
	}

	return &Source{vars}
}

// Configure creates a Source that automatically includes all Variables declared in the
// application through NewVar or the typed variants.
func Configure() *Source {
	return defaultSource
	// clone := *defaultSource
	// return &clone
}

func (s *Source) Vars() []_Var {
	return s.vars
}
