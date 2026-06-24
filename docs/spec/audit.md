# audit

- **Status:** implemented
- **Date:** 2026-06-18 (proposal); implemented 2026-06-24
- **Origin:** TIES task 4 ported sso-control's audit trail near-verbatim. Two
  services now carry the same code. See `tie/docs/notes/audit-findings.md`.
- **Implementation:** `app/audit/` — ported verbatim from `prod9/tie` `api/audit`
  (`event.go` + migration), minus TIES-specific action constants; migration re-stamped
  `202606241830`. Builds/vets clean; the DB-backed funcs need `DATABASE_URL` to exercise.

## Summary

Bake the audit trail into fx so services stop copy-pasting it. The write/read
data layer (`Record`/`Log`/`List`/`Actor`/`Event`) and the `audit_events`
migration are fully service-agnostic and map cleanly onto the existing
`app/settings` fragment shape. Ship those; leave action constants, actor
construction, and the read endpoint to callers. This captures ~90% of the
duplication with zero new coupling and **no fx core changes** — the migration
machinery already supports it.

The one structural wrinkle: the package wants to be both a *pure-function*
import (`audit.Log`) and a *migration carrier*, and in fx those are two
different things wired through two different mechanisms. See the blocker.

## What ships

| Concern             | Ships in fx?    | Where                                          |
| ------------------- | --------------- | ---------------------------------------------- |
| `audit_events` SQL  | **Yes**         | `app/audit/*.sql`, embedded on the fragment    |
| `Record` / `Log`    | **Yes**         | `app/audit/event.go` (pure funcs over `data`)  |
| `List`              | **Yes**         | `app/audit/event.go` (over `data/page`)        |
| `Actor` / `Event`   | **Yes**         | `app/audit/event.go`                           |
| Action constants    | **No** — caller | service owns its `entity.verb_past` set        |
| Actor construction  | **No** — caller | two-field projection from the session user     |
| Read HTTP endpoint  | **No** — caller | gating differs per service (see Q3)            |

## The four open questions, answered against fx

### Q1 — Migration ownership: can fx ship the migration as a built-in?

**Yes, with no fx changes.** This is already how `app/settings` and `app/files`
ship their schema. The mechanism:

- The fragment declares `//go:embed *.sql` and calls
  `.EmbedMigrations(migrations)` on its `app.Builder`
  (`app/settings/settings.go:12`, `app/files/files.go:15`).
- `app.Start` → `embedMigrations` recurses through mounted children and calls
  `migrator.Embed` for each fragment's `EmbeddedMigrations()`
  (`app/app.go:47`). `migrator.Embed` **accumulates** sources — fragments
  coexist, none overwrites another (`data/migrator/source.go:90`). This is the
  v0.8.5 behaviour the findings doc refers to; pre-0.8.5 last-mount-wins does
  not apply.

So an adopting service writes:

```go
app.Build().Name("myapp").AddDefaults().
    Mount(audit.App).   // ← runs audit_events migration, nothing to hand-copy
    Mount(ties.App).
    Start()
```

The `audit_events` table and its `(occurred_at DESC, id DESC)` index move into
fx verbatim. **Recommendation: ship it as a fragment migration, mirroring
`settings.App`.** No new fx migration machinery is required.

Caveat — dev-mode disk-shadows-embed: `migrator.LoadAuto` skips embedded
sources entirely if the CWD has any `*.up.sql` on disk
(`data/migrator/source.go`). A service that keeps its own migrations on disk in
dev runs from disk and will **not** see the embedded `audit_events` unless it
also collects fragment migrations onto disk (`data collect-migrations`) or runs
from a clean tree. This is a pre-existing fx gotcha, not new to audit, but it
bites audit harder because the SQL lives in a dependency, not the service repo.
Call it out in the adoption guide.

### Q2 — Actor extraction: caller-constructed vs context hook?

**Keep `Actor` caller-constructed (option a).** fx has no session or user
concept — that is deliberately per-service (the skill's controller examples call
a service-provided `auth.UserFromRequest`). A `audit.ActorFromContext(ctx)` hook
would force fx to define a registration seam for something it otherwise knows
nothing about, and buys little: the coupling it removes is a two-field
projection (`{ID, Email}`) the caller already has in hand at the mutation site.
`Actor` stays a plain value; the zero value remains the system actor. **No
fx-side auth dependency, no global hook state.**

### Q3 — Read endpoint: mountable controller vs per-service?

**Leave it per-service; ship only the data layer.** The two existing consumers
already diverge on gating — sso-control gates on an `audit.view` permission,
TIES on a bare session (single-tier). A shipped `audit.Controller` would have to
either bake in a gating model fx can't know, or mount ungated and rely on each
service wrapping it in the right middleware — the latter is a footgun (mount it,
forget the guard, leak the trail).

`List(ctx, page.Meta)` is the whole reusable surface; the handler around it is
~6 lines (`page.FromRequest` → `List` → `render.JSON`). That is not worth a
coupling seam. **Optionally** ship a `ListHandler` *function* (not a mounted
controller) services can wire into their own gated route — opt-in, no implicit
mount, no gating assumption. Recommend deferring even that until a third
consumer asks.

### Q4 — Detail typing: `any` → JSONB, or typed payloads?

**Keep `detail any` → JSONB.** It has served both services unchanged. `Record`
already `json.Marshal`s and writes `$6::jsonb`, and `Event.Detail` reads back as
`json.RawMessage` — callers type it at the edges where they know the shape. A
typed-payload helper (generics over a `Detail` interface) adds API surface for a
problem neither consumer has hit. If a service wants type safety it defines its
own payload struct and passes it as `any` — marshalling is transparent.

## Considered and rejected: builder-configured actors/actions

Explored (2026-06-24) making `App` a constructor taking config —
`audit.App(audit.WithActor(...), audit.AllowedActors("system","user","admin"))`. Rejected
on both axes:

- **A runtime allow-list (`AllowedActors`/`AllowedActions`)** is strictly weaker than
  caller-side `const`s: a const typo fails `go build`; an allow-list only fails when the
  `Log` line runs — and audit's non-fatal swallow means it can at most warn, never reject.
  Constants dominate it. It's also a registry (principle #1 steers away), and doesn't even
  fit `Actor{ID, Email}` — there's no role field to validate.
- **`WithActor(func(ctx) Actor)`** removes the per-call actor argument but only on request
  paths (needs the fragment's middleware), forcing a second `LogAs` for workers/CLI — more
  surface for a small win, edging toward the config #1 avoids.

Conclusion: keep `var App` plain, `Actor` caller-constructed (Q2), actions caller-owned
constants. Document the `entity.verb_past` convention; don't encode it.

## Package layout

Lives at `app/audit/` — it is an app fragment (carries a migration + optional
handler), so it belongs under `app/` next to `settings` and `files`, not at the
repo root next to leaf packages like `cache` or `secret`.

```
app/audit/
  audit.go                          # var App = app.Build().EmbedMigrations(migrations)
  event.go                          # Actor, Event, Record, Log, List
  202606181300_create_audit_events.up.sql
  202606181300_create_audit_events.down.sql
```

`event.go` moves over verbatim from TIES minus the service-specific action
constants (those stay in the service). The TIES file already imports only
`fx.prodigy9.co/data`, `data/page`, and `fxlog` — all fx-internal, so the move
introduces no new dependency.

## Public API surface

```go
package audit

// Actor is the operator behind an action. Zero value = system actor (NULL cols).
type Actor struct {
    ID    int64
    Email string
}

// Event is the stored row, the read model returned by List.
type Event struct {
    ID         int64           `json:"id"          db:"id"`
    OccurredAt time.Time       `json:"occurred_at" db:"occurred_at"`
    ActorID    *int64          `json:"actor_id"    db:"actor_id"`
    ActorEmail string          `json:"actor_email" db:"actor_email"`
    Action     string          `json:"action"      db:"action"`
    TargetType string          `json:"target_type" db:"target_type"`
    TargetID   *int64          `json:"target_id"   db:"target_id"`
    Detail     json.RawMessage `json:"detail"      db:"detail"`
}

// Log records an event, swallowing+logging any error (post-success, non-fatal).
func Log(ctx context.Context, actor Actor, action string, targetID int64, detail any)

// Record writes one row and returns its error (for the path that asserts).
func Record(ctx context.Context, actor Actor, action string, targetID int64, detail any) error

// List returns audit rows newest-first, paginated.
func List(ctx context.Context, pm page.Meta) (*page.Page[*Event], error)

// App is the mountable fragment carrying the audit_events migration.
var App = app.Build().EmbedMigrations(migrations)
```

The action *string* convention (`entity.verb_past`, `target_type` derived from
the prefix in `Record`) is documented but not enumerated by fx — services define
their own constants.

## What in fx would need to change

**Nothing required.** The migration ships through the existing fragment path;
the data funcs use already-exported `data` / `data/page` / `fxlog` surface.

Optional, only if Q3's `ListHandler` is wanted later: a thin exported func in
`app/audit` — still no core change.

## Biggest blocker / risk

**Use requires both `import audit` and `Mount(audit.App)`, and forgetting the
Mount fails silently.** The package is dual-natured: `audit.Log` is a pure
function you call from mutation sites (needs only the import), but the
`audit_events` table arrives only via the migration bound to `audit.App`, which
ships only if the service `Mount`s the fragment. These are two separate wiring
acts through two separate fx mechanisms (`import` vs `app.Builder.Mount` →
`migrator.Embed`), and nothing ties them together.

Failure mode: a service imports `audit`, calls `Log` at every mutation, compiles
clean, and ships — but never mounts `audit.App`. There is no `audit_events`
table; every `Log` swallows its insert error by design (post-success,
non-fatal), so the trail is silently empty in production with no error, no
panic, no failed request. The disk-shadows-embed gotcha (Q1 caveat) compounds
this: it can also make the table appear in dev but vanish in a clean-CWD prod
image.

Mitigation, not a fix: document "mount `audit.App`" as mandatory in the adoption
guide, and consider a cheap startup probe (e.g. audit's first `Record` checks
the table exists once and `fxlog.Warn`s loudly if absent). A true fix —
coupling the import to the migration — would need an fx seam that doesn't exist
and arguably shouldn't (it would make every imported package able to inject
schema). Accept the two-step wiring; defend it with docs and a loud-on-missing
probe.
