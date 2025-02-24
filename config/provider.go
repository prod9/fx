package config

type Provider interface {
	Initialize() error

	// TODO: Change from `string` to `[]byte` to support more complex configuration values.
	Get(name string) (string, bool, error)
	Set(name string, val string) error
}

func DefaultProvider() Provider {
	if defaultSource.provider == nil {
		defaultSource.provider = EnvProvider{}
	}
	return defaultSource.provider
}

func SetDefaultProvider(provider Provider) {
	defaultSource.provider = provider
}
