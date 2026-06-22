# prompts

- **Status:** implemented
- **Date:** 2026-06-21
- **Origin:** Studied Rust's `inquire` as a capability bar and re-examined our prompt
  package against it. "Like inquire" is a quality bar, not a feature port.

## Summary

`cmd/prompts` is a ~180-line wrapper over `pterm`'s three interactive widgets. Its real
value is not the widgets but the **args → TUI → bail fusion**: every method first
consumes a leftover CLI arg, else runs an interactive prompt when on a TTY-and-not-CI,
else fails fast. One command is scriptable, CI-safe, and interactive at once. The widgets
under it are interchangeable.

This spec: **drop `pterm`, hand-roll the TUI on `golang.org/x/term`**, keep the package
at `cmd/prompts`, preserve the fusion and every current method, and add the capabilities
that are real gaps — validation, `MultiSelect`, password-confirm, filterable/paginated
select. Reject the heavyweight `inquire` surface (calendar, `$EDITOR`, autocomplete,
theming, `CustomType<T>`) until a downstream command earns it.

## Location: stays `cmd/prompts`

Interactions belong in `cmd/` — prompts exist to serve cobra commands and nothing else
should reach for them. Keeping it under `cmd/` is the architectural statement that
interactive IO is a command-layer concern, not a library primitive. No import-path
change, no call-site churn; the seven in-tree callers are untouched.

## Why hand-roll, and on what

`x/term` is the floor, and it is effectively free:

- It is **already in the tree** (`v0.33.0`, pulled transitively via `cmd/prompts`→
  `pterm`). Its only dependency, `x/sys`, is **already required** by
  `httpserver/middlewares` and `sentry-go`. Promoting it to direct adds no module.
- It is **not a TUI framework** — its surface is `MakeRaw` / `Restore` / `GetState` /
  `GetSize` / `ReadPassword`. Exactly the "raw-mode + helpers" floor.
- Below it is only `x/sys/unix` raw termios, which means owning the
  macOS/Linux/Windows platform branches ourselves — more code to avoid a free dependency.

We use `x/term` **only** for `MakeRaw`/`Restore`/`GetSize` (raw mode is needed only for
the menu prompts and masked password). Everything above we hand-roll dep-free: ANSI
escape strings, a key decoder, a small redraw loop.

### TTY detection stays on `mattn/go-isatty`

`term.IsTerminal` has had reliability problems in practice; `go-isatty` is the trusted
check and stays the dependency for `New`'s interactivity test. We deliberately do **not**
collapse it into `x/term`.

### The deletion (philosophy G1)

Dropping `pterm` removes a subtree that routes *only* through this package:

| Module                                                          | Why it leaves   |
| -------------------------------------------------------------- | --------------- |
| `github.com/pterm/pterm`                                        | the widget lib  |
| `atomicgo.dev/{keyboard,cursor,schedule}`                       | pterm-only      |
| `github.com/gookit/color`, `github.com/xo/terminfo`            | pterm-only      |
| `github.com/lithammer/fuzzysearch`                             | pterm-only      |
| `github.com/containerd/console`, `github.com/mattn/go-runewidth` | pterm-only    |

**Net: ~8 modules removed, zero added** (`go-isatty` stays, `x/term` was already there).
The simplification is the proof we chose the right primitive.

## API — no builders, stay minimal

Builders (`p.Text("x").Default(d).Filter().Ask()`) are out — un-Go-like. Every current
positional method stays exactly as the call sites use it:

```go
p.Str("name")
p.OptionalStr("region", "ap-southeast")
p.SensitiveStr("password")
p.YesNo("proceed")
p.List("which dir", def, options)
p.OptionalList("which dir", def, options)
prompts.GenList(p, "env", def, envs, namer)
```

The only addition is one new method in the same positional style:

```go
p.MultiSelect("which tables", options, defaults) // []string; defaults pre-checked
```

No validators (the package has none today and gains none), no options structs, no
variadic config. There is nothing to attach.

## Scope — basic UX, not inquire parity

We are **not** matching inquire's breadth. No new primitives chasing it.

- **Reimplement (parity, swap pterm→x/term):** Text, masked Password, Confirm/YesNo,
  single Select (arrow-key), the Optional* variants, GenList.
- **One addition:** `MultiSelect(question, options, defaults)` — the single common
  missing type; `defaults` pre-checks entries; returns `[]string`.
- **Rejected — do not add:** validators (there are none today), substring/fuzzy filter,
  DateSelect (calendar), Editor (`$EDITOR`), autocomplete, `CustomType<T>`,
  theming/RenderConfig. Each waits for a real downstream need.

## Internals (the simplification target)

- `cmd/prompts/ansi.go` — escape-code constants + helpers (cursor up/hide/show, clear
  line, erase-down, color wrappers honoring `NO_COLOR`). ~40 lines, no deps.
- `cmd/prompts/tty.go` — `x/term` wrapper: enter/exit raw mode, terminal size, `readKey`.
- `cmd/prompts/key.go` — byte → key decoder (enter, arrows, backspace, space, esc,
  ctrl-c, rune); handles `ESC [ A/B/C/D` sequences.
- `cmd/prompts/prompts.go` — the fusion (args/CI/ALWAYS_YES) + the public methods.
- `cmd/prompts/input.go` — cooked-mode line + confirm + masked-secret readers.
- `cmd/prompts/menu.go` — raw-mode arrow-key Select / MultiSelect rendering.

TUI renders to **stderr** so stdout stays clean for piping. Unix terminals are the
target; Windows is best-effort (modern VT terminals work via `x/term`'s `MakeRaw`).

## Out of scope / deferred

Theming, Windows-legacy console, fuzzy (vs substring) filtering, and the rejected
`inquire` widgets above. Each is a cheap addition when a downstream command needs it —
narrated here so the seam is known, built on demand.
