// Package buildinfo tells the running program its name and version. See
// docs/spec/buildinfo.md.
package buildinfo

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
)

type (
	Info struct {
		Name     string
		Version  string
		Revision string
	}

	contextKey struct{}
)

// Get returns what Go stamped into the binary, named after the binary's file.
func Get() Info {
	info := Info{Name: filepath.Base(os.Args[0])}
	build, ok := debug.ReadBuildInfo()
	if !ok {
		return info
	}

	info.Version = build.Main.Version
	for _, setting := range build.Settings {
		if setting.Key == "vcs.revision" {
			info.Revision = setting.Value
		}
	}
	return info
}

func NewContext(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}

func FromContext(ctx context.Context) Info {
	if info, ok := ctx.Value(contextKey{}).(Info); ok {
		return info
	}
	return Get()
}

func (i Info) WithName(name string) Info {
	i.Name = name
	return i
}
