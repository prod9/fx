# TODO

Running list of follow-ups, design rethinks, and known-but-deferred work. Not
permanence-classified — items move out when they ship (commit, spec doc, decision
record) or get dropped explicitly. Sits next to `spec/`, `decisions/`, `notes/`
rather than inside them.

## Open

### Worker/migrator "subsystem" refactor (0.9.1 / 0.10) — 2026-07-09

Deferred from the seeding work. `ensureJobsTable` now runs at the jobs-access entry
primitives (`62dd987` + the `takeOnePendingJob` guard), so a job can be scheduled or polled
from any context on a fresh DB. Three larger threads are parked for a dedicated pass:

- **Explicit subsystem mounting** — a `subsys` concept for cross-cutting infra (migrator,
  worker) that registers its own initialization and is mounted *explicitly*, not magically:
  `app.Build().UseMigrator().UseWorker()`. Distinct from app fragments/slices.
- **App-lifecycle hook** — possibly a plain `OnStart` (chakrit not fully convinced yet).
- **Revisit the jobs-table guard placement** — the entry-point vs interior split (guard
  `scheduleJob`/`findPendingJobByName`/`takeOnePendingJob`, not the `mark*` helpers) isn't
  obvious from reading the code, and `worker.Start`'s explicit `ensureJobsTable` is now
  redundant with the poll-path guard (keep it as subsystem self-init, or drop it). Settle
  all three together.

### `app/files` audit — SHIPPED v0.9.0 (2026-07-10)

**Released v0.9.0** — tag `805eefd` on `gh` at commit `4e1b317`, consumable via
`go get fx.prodigy9.co@v0.9.0`. The whole audit + cleanup-rework arc landed. Cut via
`./platform release --minor` (CHANGELOG cut to `## v0.9.0` in `4e1b317`). Audit surface is
settled — next `/ace` should NOT re-audit.

**Release-mechanics slip (fixed).** First `./platform release v0.9.0` mis-tagged **v0.8.8**:
the pinned `platform` (`v0.8.2`, see `./platform`) silently ignores the positional `(name)`
arg under the `semver` strategy and patch-bumps instead. v0.8.8 (a previously-yanked number)
got pushed to `gh`, then deleted (remote + local) — proxy hadn't scraped it, no harm.
Re-cut with `--minor` → v0.9.0. Guardrail landed in `docs/spec/releasing.md` (commit
`a3d8ad8`, **unpushed** — awaiting chakrit's push): never pass a positional version, always
`--patch`/`--minor`/`--major`.

**Historical detail (pre-ship state, kept for the record):** unpushed arc on `main` was
`5f5d943..c0364bf`; tree clean; gofmt/build/vet + full test suite green (DB-gated tests run
off the committed `.env` DATABASE_URL).

**2026-07-10 session (`/ace-audit` before v0.9.0):** cut the invented `coda` client task (no
such client existed). Cut `blobstore.ForceDeleteObject` + `options.setDefaults` (dead). Then
the cleanup subsystem was reworked end to end — see the **Superseded 2026-07-10** note below:
`ListObjects` cut, `ObjectExists` + blobserver HEAD added, `runCleanup` now DB-driven and
fail-fast, `ScheduleCleanup` replaced by exported `CleanupJob` (sendfile seeds it in `main`),
`errutil.Aggregator` documented, `TestRunCleanup` added. **Next step: push + tag v0.9.0**
(`./platform release --minor`). Batch handling in the sweep is the one deferred follow-up.

Audit fixes landed: `multiFileCtr.Destroy` status mapping + redundant `string(Kind)` drop
(`d7cf0de`); `app-fragments.md` refreshed to the 0.9 API + exclusive-bucket warning
(`e84c988`). Struck the Mode-bitmask finding again — chakrit: it's fine, don't re-flag
(user memory `feedback_files_mode_bitmask_ok`). Kept `WithMode`/`WithOwnerIDFunc` as
functional options (not an Options struct).

**Second re-audit 2026-07-09** (a fresh `/ace-audit` pass) caught what the first missed:
`singleFileCtr.Destroy` still returned **500 on `IsNoRows`** where its multi-file twin
(fixed in `d7cf0de`) and every read handler map to 404 — the "fixed one of a pair" gap.
Fixed (`e7e1739`). Landed follow-ups:

- **Cleanup memory scoping** — `runCleanup` was an unbounded `SELECT * FROM files` + full
  bucket listing every hour. Reworked to a record-driven window: `SELECT … WHERE created_at
  > now-2·deadTimeout`, `idFloor = min(row id)`, and objects with a path-id below the floor
  are skipped (settled by earlier sweeps — the induction holds because sweeps run regularly).
  Cadence dropped `1h → deadTimeout/2` (12h). mtime rejected as an ordering key (mutable — a
  stray touch/rsync/restore breaks it); the immutable path-id drives it instead.
- **`_getOwnerID` → `defaultOwnerID`** — the underscore had no collision reason (package func
  vs struct field don't clash in Go); carried over from the original port.
- **Retracted:** the `blobstore.Client.bucket` "race" — `bucket` is written once under the
  init write-lock and every read passes through `getMinio`'s RLock first, so it's safely
  published; `-race` would not flag it.

**Superseded 2026-07-10 — cleanup reworked to DB-driven probing.** The windowing + id-floor
above is gone. `runCleanup` now iterates recent `files` rows and probes each object directly
via `blobstore.ObjectExists` (new; minio `StatObject`, blobserver gained a HEAD handler),
pruning only rows whose object never landed — it never enumerates or deletes from the
bucket. This retires the whole-bucket `ListObjects` (cut from `blobstore` + blobserver as a
`ListObjectsV2` pagination footgun) and the below-floor residual: orphan objects (object
with no row) are now **out of scope by design**, not a deferred reclaim — a failed `Destroy`
object-delete leaks permanently, judged acceptable (the row-first upload flow makes it
rare). Fail-fast on the first probe/delete error (all systemic), so a broken store can never
produce a batch-sized error storm. `runCleanup` now has end-to-end coverage (`TestRunCleanup`,
blobserver + test DB). **Deferred:** batch handling in the sweep (fine for now).

Untouched nits from the same pass (not landed): example/CLI (`download_cmd` truncates dest
on mid-copy failure; `sendfile/drops/ctr.go` magic `500` + `getDropID` sentinel-`0`
collapsing DB-error/not-found).

**Cleanup-worker seeding — resolved.** `files.ScheduleCleanup` removed; the job instance is
exported as `files.CleanupJob` and callers seed it directly with
`worker.ScheduleNowIfNotExists(ctx, files.CleanupJob)` (sendfile `main` does this). Backed by
`ensureJobsTable` running at the jobs-access entry primitives (`62dd987`, `d47a600`), so a
fresh DB seeds cleanly from any context. Ergonomic auto-seeding remains **deferred to the
subsystem refactor** (TODO item above).

**Next `/ace` — v0.9.0 is shipped. Remaining open items:**
(1) push commit `a3d8ad8` (releasing-spec guardrail) — waits on chakrit's say-so;
(2) propose the `prod9-fx` school skill update (see follow-up below) — release has landed, so
now actionable via `ace-school`;
(3) scope the subsystem refactor (incl. cleanup batch handling, deferred this session).

Follow-up (now actionable — release landed): the **`prod9-fx` school skill** documents the
old files API (`files.NewApp(client)`, `WithClient`, `WithLinkAge`) — all removed in v0.9.0;
propose a skill update via `ace-school`. Original plan below, for the record:


Audit of `app/files` (ported from `bluepages/api/files` in `bfd7081`, 2026-03-04, Chakrit
+ Opus 4.6 — the `With*`/`NewApp(client)` parameterization was added *during* that port,
not verbatim bluepages, not this-session Opus). chakrit's call: fix everything below,
including the builder redesign; bump 0.8 → **0.9.0** since the controller API breaks.
Walking via `1-by-1`; decisions collected before any edit.

Findings (audit numbering):
1. **Single-file kind doesn't enforce single.** `create_file.go:47` unconditional INSERT;
   `single_file_ctr.go:76` never deletes prior row/object → repeat upload orphans the old
   object in the store. Real data-loss-adjacent leak.
2. **`DestroyUniqueFile` mishandles multi-row** (`file.go:131`): `DELETE … RETURNING *`
   into single-row `data.Get`; with accumulated rows (from #1) all DB rows deleted but only
   one object removed → orphans. #1+#2 are one fix: single-file replace-on-create semantics.
3. **`NewApp(client)` ignores its param** (`files.go:27`, `_ = client`) — misleading;
   constructs no controllers, only embeds migrations. Drop the param or actually wire it.
4. **`WithKind` redundant** (`base_ctr.go:55`) — kind already positional to `Controller`.
5. **No tests** for the whole package (mode gating, owner resolution, single/multi, presign,
   S3 delete). Bugs #1–#2 live in the untested surface.
6. **Three client mechanisms, one fake** — global `STORAGE_URL`, per-ctr `WithClient`, dead
   `NewApp(client)`; `if client != nil … else global` repeated 3× in `file.go`. Consolidate.
7. **Metadata client-asserted, never verified** — `content_length`/`content_type` trusted
   into DB in `CreateFile`; no post-upload reconcile, no size cap. Document the trust model.
8. **`Mode` constants untyped** (`files.go:39`) — `type Mode uint8` declared but
   `ModeReadOnly`/`ModeReadWrite` are untyped ints in the iota block.
9. **Two link-age fields** — `baseCtr.linkAge` + `{single,multi}FileCtr.linkAgeCfg`; relies
   on value-receiver mount-mutation persisting into bound method values. Collapse to one.
10. **`With*` functional-options → plain funcs / options struct** (the big one) — matches
    the standing `docs/TODO.md:25` "builder/fluent un-go-like" preference. Breaking; drives
    the 0.9.0 bump. Redesign the `Controller(kind, …Option)` surface.

Cross-cutting: 0.9.0 CHANGELOG + release; downstream (`files.App` consumers, sendfile
example) updated to the new controller API.

### fx-wide error conventions audit — 2026-07-09

Audit the whole codebase for error-handling conventions and establish a written standard,
then align packages to it. Questions to settle:
- When to define a sentinel (`errors.New`/`Err*`) vs decorate with `errutil.WithCode`/
  `WithData`/`Wrap` vs a `validate` field error — no consistent rule today.
- How HTTP status is chosen. `render.Error` takes `status` as an explicit arg
  (`httpserver/render/render.go:34` inline TODO already flags this breaks SRP — the
  originating error should carry its code, not the controller). Fold that TODO in here.
- Error `code`/`description` consistency across packages (`httperrors` constants vs
  ad-hoc `fmt.Errorf` strings vs coded errors).

Output: a decision/spec fixing the convention, then a sweep aligning packages. Surfaced
2026-07-09 during the `app/files` redesign — the files `MaxSize` rejection needs a
well-defined coded error, which exposed the absence of a codebase-wide rule.

### Prompts redesign shipped as v0.8.6 (2026-06-22) — two loose ends

`cmd/prompts` reimplemented hand-rolled on `golang.org/x/term`; dropped `pterm` + ~10
subtree modules (`go-isatty` kept — `term.IsTerminal` is unreliable per chakrit). Added
`MultiSelect(question, defaults, options)` + `OptionalMultiSelect` (arg order mirrors
`List`/`OptionalList`; `defaults` pre-check, base bails non-interactive). Spec:
`spec/prompts.md` (Status: implemented). Released **v0.8.6** (tag `4fd53f3`); platform
integrated, pinned `fx.prodigy9.co@v0.8.6`, dropped its `replace => ../fx`.

Resolved 2026-06-24: pushed `64b3e4d` + the doc commits to `gh/main`; `spec/audit.md`
committed as a tracked draft spec. The builder/validation alternative for audit
actors/actions was explored and rejected (kept `var App` + caller constants; rationale in
`spec/audit.md`).

Still open:
- Low-pri school candidate: propose to `go-coding` that builder/fluent APIs are out
  (chakrit: "un-go-like") — prefer plain funcs + positional args or a plain options struct.

### Audit app (`app/audit`) — shipped v0.8.7, one verification open

`app/audit` fragment (`audit.go` + `event.go` + migration `202606241830`) **ported
verbatim** from `prod9/tie` `api/audit` minus TIES action constants. `Record`/`Log`/`List`/
`Actor`/`Event`; `var App = app.Build().EmbedMigrations(...)`, no controller. Actions/actors
are caller-owned constants. No dual-wiring probe — `Log` swallows-and-logs, so a missing
table warns on every call. **Released v0.8.7** (tag `f69d3fb`); tie adopted it back as a
library import, build/vet/test green (its DBs are resettable, so no ledger surgery).

- **Open (chakrit's):** runtime `data migrate` + `Record`→`List` against Postgres, after a
  tie reset — the only DB-backed exercise nobody could run (no `DATABASE_URL` in sandbox/CI).
- Migration identity: fx owns its timestamp; consumers reconcile downstream (reset or
  resync). Adoption note in `spec/audit.md`; currently `## Unreleased` in CHANGELOG.
- **v0.8.8 was cut docs-only then YANKED** (tag deleted, proxy 404, tie never pinned it).
  Guardrail added to `releasing.md`: docs-only changes don't warrant a release. The
  Unreleased adoption note rides the next code-bearing release.

### Philosophy promotion — greenlit, still unexecuted

G1/G2 → `spec/philosophy.md`, fold G3/G4/G6 (see the older TODO entry below). chakrit
greenlit "your call on all items"; not yet done — top backlog item.

### ace-connect bridge live — autonomous, release-cutting held

Listener bound: slug `prod9.fx.claude`, autonomous mode (the Monitor survives `/clear` —
next `/ace` should load `ace-connect` to recover wire format/mode). Envelope: safe
bugfix-class work + commits on `main` proceed; pushes, releases, deploys wait for chakrit.
No `.inbox.log` (autonomous acts directly). This session it handled platform's MultiSelect
pre-check request end-to-end over the bridge.

### Revealed-philosophy analysis — promotion decision pending (2026-06-21)

Distilled how FX was actually built from the pre-AI git history (174 commits through
`bde86f6`) and checked it against the eight stated principles in `spec/philosophy.md`.
Full write-up: `notes/2026-06-21-revealed-philosophy-from-git-history.md`. Result: all
eight are authentic (several written into doc comments predating CLAUDE.md), plus six
unstated gaps grounded only in the spine. (Authorship caveat: subtree/squash imports
attribute every pre-AI commit to chakrit; some leaf code is from others — don't infer
authorship from git metadata.)

Open decision (chakrit's call): whether to promote into `spec/philosophy.md` —
- **G1** subtraction-via-stronger-primitives — "net deletion is the signal you chose
  the right primitive"; the worker's Postgres-CAS-over-`FOR UPDATE` decision is the
  worked example, and it also grounds stated principle #7.
- **G2** provenance + design-in-prose-build-on-demand (the `config.Provider` seam
  narrated at birth in `config/source.go`, built 2.5y later in `8ad5c6d`).
- G3 (ship-rough-then-harden), G4 (caller-safety over author-convenience), G6
  (anti-dogmatism) are lower-altitude; fold into existing principles if promoted.

**Greenlit 2026-06-22** — chakrit: "your call on all items, I'm fine with it."
Authorized but not yet executed (this session went to the prompts redesign + v0.8.6
release instead). **Top pending task next session: execute the promotion** — promote
G1 + G2 (as new principles or a clearly-marked section), fold G3/G4/G6 into existing
principles, per the note's own recommendation.

*Logged: 2026-06-21.*

### Triaged 2026-06-16 — four small TODOs walked, three decisions

Walked the inline TODOs marked "small / well-scoped" in a 1-by-1 session.
Decisions captured; no code yet. Execution order below is a recommendation, not a
commitment.

1. ~~**`fxlog`: hoist process termination out of `Sink`.**~~ Shipped 2026-06-17.
   Dropped `Fatal` from the `Sink` interface; added optional
   `Flusher{ Flush() error }`. Package-level `fxlog.Fatal` is now the single owner
   of "log → flush if supported → `os.Stderr.Sync()` → `os.Exit(1)`". `ZerologSink`
   no longer uses zerolog's inline `.Fatal()` builder. Resolved
   `fxlog/slog_sink.go:35`.

2. ~~**`Home` readiness probe (`/healthz`).**~~ Shipped 2026-06-16. Added
   `data.LookupFromContext` (stdlib-style comma-ok sibling of `FromContext`)
   and `Home.Healthz` at `/healthz` with the 500ms dep-reachability behavior
   from `notes/2026-06-16-readiness-probe-semantics.md`.

3. **Context-threading rethink** (folds in `app/settings/provider.go:57` and
   `app/settings/settings.go:42`). Both are symptoms of a larger design gap:
   how should `*config.Source`, `*sqlx.DB`, and `context.Context` thread
   through non-HTTP code paths — settings fragment, workers, CLI subcommands,
   init paths? The `Provider.dbContext()` helper rebuilds `context.Background()`
   per call because there's no clear answer to "what context do I belong to
   when there's no caller ctx?" And the settings cache shape (`Get(ctx, key)`
   TODO) can't be decided independently of where the long-lived Provider
   state lives. Output: a written design proposal in `docs/notes/` or a
   decision in `docs/decisions/`, not code. When the cache is implemented as
   part of the redesign, use the existing `cache` package (extend if needed)
   rather than inventing settings-local cache state.

Migrator merge (existing entry below) was deferred separately — chakrit's call
that "merging may not be the right thing to do" and the question wants longer
think-time before any code change.

### Watch for `prod9-fx` skill update from school

`prod9.school.claude` is baking claims 1, 2, 4, 5, 6, 7 from
`/tmp/fx-verdicts-prod9.fx.claude.md` into the skill (with app-level framing on
claim 6). Skim the resulting PR/commit when it lands to make sure the framing
matches fx behavior.

Update 2026-06-08: confirmed v0.8.5 boundary with school over ace-connect.
- Claim 1 framing: "Mount auto-aggregates fragment migrations >=0.8.5; copy-in-tree
  is the legacy <=0.8.4 workaround." School baking, citing 0.8.5 / 61a66d3.
- Claim 2: clean-CWD-to-exercise-embed note folded in (the disk-vs-embed footgun
  one above).

### `migrator.LoadAuto`: merge disk + embed, or rethink loading

`data/migrator/source.go:110-150` — current behavior is disk-OR-embed,
mutually exclusive. If CWD scan finds any `*.up.sql`, embedded migrations are
skipped entirely. App-global precedence.

The footgun: dev loops on disk SQL, so any embedded fragment migrations
(`files.App`, `settings.App`, etc.) silently don't run in dev — only in
prod-deploy when the app has no disk SQL and falls through to embed.

**Confirmed in practice 2026-06-08** while verifying v0.8.5: `examples/migrations/`
must be built and run from a clean CWD (e.g. `/tmp`) to exercise the embed path,
otherwise the disk walker recurses into `users/`, `posts/`, `comments/` and
short-circuits before the embedded `files.App` migration is even considered. See
`examples/migrations/README.md` for the build+run workaround.

Direction to consider:
- **Merge:** union disk + embed sources, dedupe by migration name/timestamp.
  Disk wins on conflict (so devs can override an embedded migration locally).
- **Rethink:** the disk-vs-embed distinction is a deploy-shape concern; maybe
  the migrator should take an explicit ordered list of sources and let the
  app/cmd decide precedence per-environment.
- **Fragmentize migrations (chakrit, 2026-07-05 — chosen next-work direction):**
  make migrations a first-class fragment-composition output. The app already
  merges controllers, commands, jobs, and middlewares deterministically up the
  `Mount` tree (`app.collect`); migration *sources* should ride the same merge
  instead of being handed off to `migrator.LoadAuto`'s disk-vs-embed heuristic.
  Then aggregation is decided by app composition (predictable), and the migrator
  just runs whatever ordered set the app hands it — no CWD-scan short-circuit,
  no clean-CWD workaround. Supersedes the earlier "merging may not be right,
  needs think-time" hold: the fix moves up a layer rather than patching
  `LoadAuto` in place.

Surfaced again 2026-07-05 building `examples/sendfile` (files.App + blobserver):
`go run . data migrate` from the example dir applied only the local `drops`
migration and silently skipped the embedded `files` table — same footgun. The
example README documents the build-and-run-from-clean-CWD path until the
fragmentize work lands.

Not urgent; flag for the next time someone hits the "my migration didn't run"
surprise.

*Logged: 2026-06-07; confirmed in practice 2026-06-08. Related: school.claude
verdicts file `/tmp/fx-verdicts-prod9.fx.claude.md` claim 2 (now baked into the
`prod9-fx` skill).*

## Inline code TODOs

Swept 2026-06-09 from `grep -rn TODO` across `*.go`. Listed by package so they can
be triaged or promoted to their own section above when picked up. Comment text is
kept verbatim for grep parity.

### `worker/`

- `worker/worker.go:76` — *"Might need to be careful with transactions here"* in
  `ScheduleAtIfNotExists`. The pending-name lookup and insert aren't wrapped in a
  transaction, so two schedulers racing on the same job name can both win.
- `worker/worker.go:246` — *"Add more speciailized errors for signaling
  retries/rerun"* on `processJob`. Today any non-nil error from `Run` is treated
  the same; no way for a job to request a retry vs. a hard fail.
- `worker/worker.go:263` — *"Enforce timeouts"* before invoking `instance.Run(ctx)`.
- `worker/worker.go:264` — *"Better to run the job in a separate transaction. So
  the job state is not effected by the job code."* Job-body work currently shares
  the worker's transactional scope.

### `httpserver/`

- `httpserver/render/render.go:34` — `render.Error` shouldn't take `status` as an
  argument; the originating error should carry it (otherwise controllers pick the
  code, which breaks SRP).
- ~~`httpserver/controllers/home.go:16` — Add a built-in `/healthz`.~~ Shipped
  2026-06-16 (`Home.Healthz`, 500ms dep-reachability probe).

### `app/settings/`

- `app/settings/provider.go:57` — *"Maybe save in struct?"* — `Provider.dbContext`
  rebuilds the `data.NewContext(...)` on every call.
- `app/settings/settings.go:42` — *"Cache"* the `Get(ctx, key)` lookup; today it
  hits the DB on every call.

### `config/`

- `config/provider.go:6` — *"Change from `string` to `[]byte` to support more
  complex configuration values."* Provider values are currently string-only.

### `fxlog/`

- ~~`fxlog/slog_sink.go:24,28` — `SLogSink.Log` / `Error` pass `nil` as context to
  `slog.LogAttrs`. `staticcheck` SA1012: use `context.TODO()` (or thread caller
  context through `Sink`, which would be a wider redesign). Pre-existing,
  surfaced during the 2026-06-17 Sink-Fatal removal audit.~~ Shipped 2026-06-17.
  Took the `context.TODO()` route; threading caller ctx through `Sink` deferred
  to the broader context-threading rethink above.

