# Build info

**Status:** implemented

Package `buildinfo` tells the running program its name and version, for error reports,
logs, and introspection routes.

## Build-time facts

`buildinfo.Get()` returns what Go stamped into the binary:

```go
type Info struct {
	Name     string
	Version  string
	Revision string
}
```

| Field      | Source                                                       |
| ---------- | ------------------------------------------------------------ |
| `Name`     | the binary's file name                                       |
| `Version`  | the main module's version from `debug.ReadBuildInfo`         |
| `Revision` | the `vcs.revision` build setting                             |

A plain `go build` in a git checkout stamps both: the tag (or a pseudo-version) as the
module version, and the commit as the revision. A build without VCS information reports
Go's `(devel)` and no revision.

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
