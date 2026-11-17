# Build info

**Status:** implemented

Package `buildinfo` tells the running program its name and version, for error reports,
logs, and introspection routes.

## Build-time facts

`buildinfo.Get()` returns an `Info` resolved once per process, read-only:

```go
type Info struct {
	Name     string
	Version  string
	Revision string
}
```

| Field      | First set of                                                                             |
| ---------- | ---------------------------------------------------------------------------------------- |
| `Name`     | link-time `name`; the binary's file name                                                 |
| `Version`  | link-time `version`; Go's module version from `debug.ReadBuildInfo`; `(devel)`           |
| `Revision` | Go's `vcs.revision` build setting                                                        |

Link-time values come from `-ldflags`:

```sh
go build -ldflags "-X fx.prodigy9.co/buildinfo.version=$(git describe --tags --always)" .
```

Go stamps the module version and VCS revision itself when the build sees the `.git`
directory, so a plain `go build` in a checkout needs no flags.

## Info on the context

There is no global name. `app.Start` puts `buildinfo.Get().WithName(app.Name())` on the
context it executes the root command with (the binary's name stays when the app has no
name), so two apps in one process each report their own name.

```go
func NewContext(ctx context.Context, info Info) context.Context
func FromContext(ctx context.Context) Info // the context's Info, else Get()
func (i Info) WithName(name string) Info
```

The context flows on from there:

* cobra hands it to every command as `cmd.Context()`.
* `serve` calls `httpserver.Server.StartContext(cmd.Context())`; the server uses it as
  `http.Server.BaseContext`, so every request context descends from it.
* `worker` calls `worker.Worker.StartContext(cmd.Context())`; job contexts descend from
  it.
* `Start()` on the server and the worker remains and calls `StartContext` with
  `context.Background()`.

CLI-only apps read `buildinfo.FromContext(cmd.Context())` in their own commands.
