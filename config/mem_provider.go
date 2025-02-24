package config

type MemProvider struct {
	values map[string]string
}

var _ Provider = &MemProvider{}

func (p *MemProvider) Initialize() error {
	p.values = map[string]string{}
	return nil
}

func (p *MemProvider) Get(name string) (string, bool, error) {
	val, ok := p.values[name]
	return val, ok, nil
}

func (p *MemProvider) Set(name string, val string) error {
	p.values[name] = val
	return nil
}
