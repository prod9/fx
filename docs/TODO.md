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

### `migrator.LoadAuto`: merge disk + embed, or rethink loading

`data/migrator/source.go:106-134` — current behavior is disk-OR-embed,
mutually exclusive. If CWD scan finds any `*.up.sql`, embedded migrations are
skipped entirely. App-global precedence.

The footgun: dev loops on disk SQL, so any embedded fragment migrations
(`files.App`, `settings.App`, etc.) silently don't run in dev — only in
prod-deploy when the app has no disk SQL and falls through to embed.
Aggregating fragment migrations into the root (Item 1 from the 2026-06-07
1-by-1) makes embedded fragments work in prod-embed mode but doesn't fix this
dev-mode invisibility.

Direction to consider:
- **Merge:** union disk + embed sources, dedupe by migration name/timestamp.
  Disk wins on conflict (so devs can override an embedded migration locally).
- **Rethink:** the disk-vs-embed distinction is a deploy-shape concern; maybe
  the migrator should take an explicit ordered list of sources and let the
  app/cmd decide precedence per-environment.

Not urgent; flag for the next time someone hits the "my migration didn't run"
surprise.

*Logged: 2026-06-07. Related: school.claude verdicts file
`/tmp/fx-verdicts-prod9.fx.claude.md` claim 2.*
