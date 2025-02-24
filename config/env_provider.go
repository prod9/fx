package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/joho/godotenv"
)

type EnvProvider struct{}

var _ Provider = EnvProvider{}

func (p EnvProvider) Initialize() error {
	envs, err := findDotEnvs()
	if err != nil {
		return err
	}

	if len(envs) > 0 {
		for i := len(envs) - 1; i >= 0; i-- {
			if err := godotenv.Load(envs[i]); err != nil {
				return fmt.Errorf("dotenv: %w", err)
			}
		}
	}

	return nil
}

func (p EnvProvider) Get(name string) (string, bool, error) {
	val, ok := os.LookupEnv(name)
	return val, ok, nil
}

func (p EnvProvider) Set(name string, val string) error {
	return os.Setenv(name, val)
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
