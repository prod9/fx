# Proposal: re-do `prompts` as a hand-rolled, dep-light TUI package

- **Status:** draft
- **Date:** 2026-06-21
- **Origin:** Studied Rust's `inquire` crate as a capability bar and re-examined our
  prompt package against it. "Like inquire" reads as a quality bar, not a feature port.

## Summary

Today's prompt code is `cmd/prompts` — a ~180-line wrapper over `pterm`'s three
interactive widgets. Its real value is not the widgets but the **args → TUI → bail
fusion**: every method first consumes a leftover CLI arg, else runs an interactive
prompt when on a TTY-and-not-CI, else fails fast. That fusion makes one command
scriptable, CI-safe, and interactive at once. The widgets under it are interchangeable.

This proposal: **drop `pterm`, hand-roll the TUI on `golang.org/x/term`**, move the
package to top-level `prompts/`, preserve the fusion and every current method, and add
the handful of capabilities that are real gaps (validation, `MultiSelect`,
password-confirm, filterable/ paginated select). Reject the heavyweight `inquire`
surface (calendar, `$EDITOR`, autocomplete, theming, `CustomType<T>`) until a downstream
app earns it — per philosophy G2.

## Why hand-roll, and on what

`x/term` is the floor, and it is effectively free:

- It is **already in the tree** (`v0.33.0`, pulled transitively via `cmd/prompts`→
  `pterm`). Its only dependency, `x/sys`, is **already required** by
  `httpserver/middlewares` and `sentry-go` regardless. Promoting it to a direct
  dependency adds no module.
- It is **not a TUI framework** — its entire surface is `MakeRaw` / `Restore` /
  `GetState` / `GetSize` / `IsTerminal` / `ReadPassword`. That is exactly the
  "raw-mode toggle + helpers" floor, not a `bubbletea`/`tcell`-style engine.
- Below it is only `x/sys/unix` raw termios, which means **owning the
  macOS/Linux/Windows platform branches ourselves** — more code, to avoid a dependency
  that costs nothing. Wrong trade.

Everything above raw-mode we hand-roll, dep-free: ANSI escape strings, a key decoder,
and a small redraw loop. Raw mode is needed *only* for the menu prompts (`Select`,
`MultiSelect`) and masked password; `Text` and `Confirm` read a cooked line (native
backspace/editing for free), which is the largest single simplification.

### The deletion (this is the point — philosophy G1)

Dropping `pterm` removes a subtree that routes *only* through this package:

| Module                                                         | Why it leaves          |
| ------------------------------------------------------------- | ---------------------- |
| `github.com/pterm/pterm`                                       | the widget lib itself  |
| `atomicgo.dev/{keyboard,cursor,schedule}`                      | pterm-only             |
| `github.com/gookit/color`, `github.com/xo/terminfo`           | pterm-only             |
| `github.com/lithammer/fuzzysearch`                            | pterm-only             |
| `github.com/containerd/console`, `github.com/mattn/go-runewidth` | pterm-only          |
| `github.com/mattn/go-isatty` (direct)                          | `term.IsTerminal` wins |

**Net: ~9 modules removed, zero added.** The simplification *is* the proof we chose the
right primitive — the worked example of the discipline the worker's Postgres-CAS
decision already demonstrates in this codebase.

## Package location

Move `cmd/prompts` → top-level `prompts/`. It is a reusable library, not a cobra
command: it depends only on `config`, `fxlog`, `slices`, and now `x/term`. `cmd/` is for
command definitions; this belongs beside `config`, `data`, `validate` in the package map.
Cost: the import path changes (`fx.prodigy9.co/cmd/prompts` → `fx.prodigy9.co/prompts`),
a flagged break in `CHANGELOG.md`. Seven in-tree call sites updated in the same change.

## API

Terse methods stay (back-compat for all 7 call sites) as thin shims; builders are the
extensible path for the richer cases. The `Session` entry point, `CI`, and `ALWAYS_YES`
are unchanged.

```go
p := prompts.New(cfg, args)

// preserved terse forms — same signatures the call sites use today
p.Str("username")                       // line input
p.OptionalStr("region", "ap-southeast") // optional trailing arg, default, no prompt
p.SensitiveStr("password")              // masked
p.YesNo("proceed")                      // p.Confirm(q, "yes", "no")
p.List("which dir", def, options)       // single select
p.OptionalList("which dir", def, opts)  // optional trailing arg, default, no prompt
prompts.GenList(p, "env", def, envs, namer) // typed select

// new builders — options without method explosion
p.Text("email").Default("x@y.co").Validate(prompts.Required).Ask()
p.Password("password").Confirm("again").Ask()        // built-in confirm pairing
p.Select("which dir", options).Default(def).Filter().Ask()
p.MultiSelect("tables", options).Min(1).Filter().Ask() // NEW: []string
```

Validation applies on the **args path too**, not just interactive input — closing a
current robustness gap where args bypass all checking. `Validator` is
`func(string) error`; ship only `Required` built-in, earn the rest.

## inquire surface: in vs out

| inquire capability     | Decision | Rationale                                            |
| ---------------------- | -------- | ---------------------------------------------------- |
| Text / line input      | **in**   | current `Str`; add default + validate                |
| Password (masked)      | **in**   | current `SensitiveStr`; add confirm-pairing          |
| Confirm                | **in**   | current `YesNo`/`Confirm`                            |
| Select (single)        | **in**   | current `List`; add filter + pagination              |
| MultiSelect            | **in**   | real gap; the main "more features" win               |
| Validators             | **in**   | thin re-prompt loop; `Required` only to start        |
| Filter on long lists   | **in**   | substring, case-insensitive; not a fuzzy crate       |
| Pagination             | **in**   | cap visible window; needed once lists are long       |
| DateSelect (calendar)  | **out**  | no downstream demand; heavy widget (G2)              |
| Editor (`$EDITOR`)     | **out**  | no demand; trivial to add when earned                |
| Autocomplete trait     | **out**  | filter covers the common case; over-engineered (G2)  |
| `CustomType<T>` parse  | **out**  | `GenList` + caller parse suffices                    |
| RenderConfig / themes  | **out**  | one built-in style; convention over options (#8)     |

## Internals (the simplification target)

- `prompts/ansi.go` — escape-code constants + helpers (cursor up/hide/show, clear line,
  erase-down, color wrappers honoring `NO_COLOR`). ~40 lines, no deps.
- `prompts/tty.go` — `x/term` wrapper: enter/exit raw mode, terminal size, `readKey`.
- `prompts/key.go` — byte → key decoder (enter, arrows, backspace, space, tab, esc,
  ctrl-c, rune); handles `ESC [ A/B/C/D` sequences.
- `prompts/session.go` — the fusion (args/CI/ALWAYS_YES) + terse methods.
- `prompts/{text,password,confirm,select,multiselect}.go` — the prompt types/builders.

TUI renders to **stderr** so stdout stays clean for piping. Unix terminals are the
target; Windows is best-effort (modern VT terminals work via `x/term`'s `MakeRaw`).

## Out of scope / deferred

Theming, Windows-legacy console, fuzzy (vs substring) filtering, and the rejected
`inquire` widgets above. Each is a cheap addition *when a downstream app needs it* —
narrated here so the seam is known, built on demand (G2).
