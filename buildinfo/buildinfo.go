// Package buildinfo tells the running program its name and version. Build-time facts
// come from Get; an Info carried on the context lets each app in one process report its
// own name. See docs/spec/buildinfo.md.
package buildinfo

import (
	"context"
	"os"
	"path/filepath"
	"runtime/debug"
	"sync"
)

const develVersion = "(devel)"

var (
	// Set at link time:
	//
	//	go build -ldflags "-X fx.prodigy9.co/buildinfo.version=$(git describe --tags --always)"
	name    string
	version string

	get = sync.OnceValue(func() Info {
		build, _ := debug.ReadBuildInfo()
		return resolve(Info{Name: name, Version: version}, build, os.Args[0])
	})
)

type (
	Info struct {
		Name     string
		Version  string
		Revision string
	}

	contextKey struct{}
)

// Get returns the build-time facts: link-time values first, then Go's build info, then
// the binary's file name and "(devel)".
func Get() Info {
	return get()
}

func (i Info) WithName(name string) Info {
	i.Name = name
	return i
}

func NewContext(ctx context.Context, info Info) context.Context {
	return context.WithValue(ctx, contextKey{}, info)
}

// FromContext returns the context's Info, else Get().
func FromContext(ctx context.Context) Info {
	if info, ok := ctx.Value(contextKey{}).(Info); ok {
		return info
	}
	return Get()
}

func resolve(linked Info, build *debug.BuildInfo, binary string) Info {
	stamped := fromBuildInfo(build)
	return Info{
		Name:     firstSet(linked.Name, filepath.Base(binary)),
		Version:  firstSet(linked.Version, stamped.Version, develVersion),
		Revision: stamped.Revision,
	}
}

func fromBuildInfo(build *debug.BuildInfo) Info {
	if build == nil {
		return Info{}
	}

	info := Info{}
	if build.Main.Version != develVersion {
		info.Version = build.Main.Version
	}
	for _, setting := range build.Settings {
		if setting.Key == "vcs.revision" {
			info.Revision = setting.Value
		}
	}
	return info
}

func firstSet(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
