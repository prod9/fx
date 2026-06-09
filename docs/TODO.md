# TODO

Running list of follow-ups, design rethinks, and known-but-deferred work. Not
permanence-classified — items move out when they ship (commit, spec doc, decision
record) or get dropped explicitly. Sits next to `spec/`, `decisions/`, `notes/`
rather than inside them.

## Open

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
- `httpserver/controllers/home.go:16` — Add a built-in `/healthz` that does more
  than the trivial home handler (e.g. ping the database).

### `app/settings/`

- `app/settings/provider.go:57` — *"Maybe save in struct?"* — `Provider.dbContext`
  rebuilds the `data.NewContext(...)` on every call.
- `app/settings/settings.go:42` — *"Cache"* the `Get(ctx, key)` lookup; today it
  hits the DB on every call.

### `config/`

- `config/provider.go:6` — *"Change from `string` to `[]byte` to support more
  complex configuration values."* Provider values are currently string-only.

### `fxlog/`

- `fxlog/slog_sink.go:35` — *"Ensure the log message is flushed before the exit"*
  in the slog Fatal path.
