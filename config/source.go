package config

import (
	"errors"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"

	"github.com/joho/godotenv"
)

const DotEnvSearchLimit = 4

var defaultSource *Source = &Source{}

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
	envs, err := findDotEnvs()
	if err != nil {
		log.Println("config:", err)
	}

	if len(envs) > 0 {
		for i := len(envs) - 1; i >= 0; i-- {
			if err := godotenv.Load(envs[i]); err != nil {
				log.Println("dotenv:", err)
				return &Source{vars}
			}
		}
	}

	return &Source{vars}
}

// Configure creates a Source that automatically includes all Variables declared in the
// application through NewVar or the typed variants.
func Configure() *Source {
	return NewSource(defaultSource.vars)
}

func (s *Source) Vars() []_Var {
	return s.vars
}

func findDotEnvs() ([]string, error) {
	var envs []string
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	} else if wd, err = filepath.Abs(wd); err != nil {
		return nil, err
	}

	for n := 0; n < DotEnvSearchLimit; n++ {
		dotenv := filepath.Join(wd, ".env")
		dotenvlocal := filepath.Join(wd, ".env.local")

		if _, err := os.Stat(dotenvlocal); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			} // else no-op
		} else {
			envs = append(envs, dotenvlocal)
		}
		if _, err := os.Stat(dotenv); err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				return nil, err
			} // else no-op
		} else {
			envs = append(envs, dotenv)
		}

		dotgit := filepath.Join(wd, ".git")
		if _, err := os.Stat(dotgit); !errors.Is(err, fs.ErrNotExist) {
			// we found a .git, most likely project's root so we stop here
			break
		} else {
			wd = filepath.Dir(wd) // go up 1 dir
		}
	}

	// reverse so it's in Overload-able order
	slices.Reverse(envs)
	return envs, nil
}
