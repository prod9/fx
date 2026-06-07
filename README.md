# PRODIGY9 FRAMEWORK

A minimalistic, modular Go API framework. Bundles well-integrated tools for building
APIs — config, routing, DB + migrations, jobs, logging, mail, caching, secrets — while
letting you swap pieces in/out.

Module: `fx.prodigy9.co` | Go 1.24

## Install

```sh
go get fx.prodigy9.co
```

Then import the packages you need:

```go
import (
  "fx.prodigy9.co/app"
  "fx.prodigy9.co/data"
  "fx.prodigy9.co/httpserver/controllers"
)
```

See [`docs/spec/`](docs/spec/) for the per-topic API reference and
[`examples/`](examples) for working applications (`todoapi`, `envfiles`, `workers`).

## Philosophy

The throughline that guides what FX is and isn't. Full version in
[`docs/spec/philosophy.md`](docs/spec/philosophy.md).

1. **Modular by composition, not by config.** Apps are built by mounting fragments
   (`app.Build().Mount(auth.App).Mount(todo.App)`) — Django-app-style bundles of
   controllers + middlewares + commands. No DI container, no plugin registry.
2. **Thin wrappers over standard primitives, never replacements.** chi, sqlx, cobra,
   `net/http` stay reachable. Add ergonomics, not opacity.
3. **Config as decentralized declarations.** Each package declares its own `config.*Def`
   next to the code that reads it. No central `Config` struct.
4. **Context as the carry-bag.** `*config.Source` and `*sqlx.DB` ride in
   `context.Context`. Transaction scopes propagate transparently.
5. **Everything is optional.** `data`, `worker`, `mailer`, `cache`, `secret`,
   `blobstore` are independent. `AddDefaults()` is convenience, not contract.
6. **Operational concerns are first-class.** `print-config`, `data migrate`, `data
   psql`, `__panic`, `MUST_MIGRATE`, embedded migrations — debuggability ships with the
   binary.
7. **Punt distributed-systems problems to infrastructure.** HA is "use pgbouncer." No
   service mesh, no retry framework, no pluggable transport.
8. **Convention over options, but one layer deep.** One recommended path plus a small
   number of escape hatches. Three transaction styles, two log sinks. Not pluggable
   everything.

Minimalism here is a discipline against accidental complexity, not a goal in itself.

## Hacking on FX from a downstream app (git subtree)

If you need to iterate on FX itself from inside an application — typically when fixing
a bug you hit downstream and want a faster loop than upstream PR + version bump — you
can pull FX in as a [git subtree][0]:

```sh
git subtree add --prefix=fx https://github.com/prod9/fx main
```

* Edit anything under `fx/` as you would your own code.
* `git subtree pull` to sync upstream changes back down.
* When you fix something, push it upstream:
  * Isolate the `fx/` changes into their own commit.
  * Prefix the commit message with `fx:`.
  * `git subtree push` back to `prod9/fx`.

This path is supported but no longer the default install. Most projects should use
`go get` and open a PR upstream if FX needs fixing.

## Vanity Server

The `main.go` at the repo root runs a tiny Go application that serves the
`fx.prodigy9.co` vanity import URL.

A docker image can be built using the `build.py` [Dagger][1] script:

```sh
pip install --upgrade dagger-io==0.6.2 anyio  # first clone
./build.py
```

[0]: https://www.atlassian.com/git/tutorials/git-subtree
[1]: https://dagger.io
