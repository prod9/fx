package fxtest

import "fx.prodigy9.co/config"

// not sure if this provider should just be placed in the config/ package directly
type testProvider struct {
	env config.EnvProvider
	mem config.MemProvider
}

var _ config.Provider = &testProvider{}

func (t *testProvider) Initialize() error {
	if err := t.env.Initialize(); err != nil {
		return err
	} else if err := t.mem.Initialize(); err != nil {
		return err
	} else {
		return nil
	}
}

func (t *testProvider) Get(name string) (string, bool, error) {
	if val, ok, err := t.mem.Get(name); err != nil {
		return "", false, err
	} else if !ok {
		return t.env.Get(name)
	} else {
		return val, ok, nil
	}
}

func (t *testProvider) Set(name string, val string) error {
	return t.mem.Set(name, val)
}

func Configure() *config.Source {
	return config.NewSource(
		&testProvider{},
		config.DefaultSource().Vars(),
	)
}
