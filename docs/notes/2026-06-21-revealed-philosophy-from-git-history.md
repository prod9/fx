# FX's revealed philosophy — what the pre-AI git history says

A study of how the maintainer actually wrote FX before any AI assistance, distilled
into higher-level philosophies and checked against the eight stated principles in
`docs/spec/philosophy.md` / `CLAUDE.md`. Question being answered: is the documented
philosophy authentic to the practice, and what does the practice reveal that the
documented list never names?

## Method and boundary

- **Corpus:** the 174 commits from `ab137b5` (2023-06-21, "Initial version") through
  `bde86f6` (2026-01-11). All 201 commits in the repo carry Chakrit as author.
- **AI boundary:** `025b773 ace-ify` (2026-03-04) — when ACE, `.claude/`, and CLAUDE.md
  entered the tree. The modern `scope:` commit convention already appears at 2026-01-11,
  ~2 months *before* Claude — so the `fx:` prefix is not the AI boundary; ace-ify is.
- **Reading around the output filter:** the lowfat hook compacts `git log`/`git show`
  output past ~25–30 lines even into a redirect, so the corpus was paged in small
  windows, and two sub-agents read the historical diffs file-by-file.
- **Authorship caveat (important).** Every pre-AI commit's author *and* committer email
  is `chakrit@prodigy9.co` — `git subtree`/squash import mechanics flattened all
  authorship to the maintainer's identity, leaving no metadata trace of the real authors.
  So "committed by Chakrit" ≠ "written by Chakrit". Some leaf code was contributed by
  others and committed under his identity — e.g. the cloudstorage/objstore client (the
  rough 2024-01-30 squashed-import cluster), whose roughness reflects nothing about his
  craft. The architectural **spine** — `app`, `config`, `data`, `httpserver`, `fxlog`,
  `worker`, `slices`, `errutil` — is his, and the findings below rest only on the spine.

## Verdict: the stated eight are authentic, not post-hoc

Every one of the eight is visibly revealed in the spine — several are written into his own
doc comments *years before* CLAUDE.md existed. The documented philosophy describes real
practice; it is not aspirational backfill.

| #  | Stated principle                  | Strongest pre-AI evidence (spine only)                                                                                |
|----|-----------------------------------|---------------------------------------------------------------------------------------------------------------------|
| 1  | Modular by composition            | `app.collect()` recursively flattens the app tree into one fragment; the interface/impl/builder split existed at birth. `httpserver/fragment.go` doc: modules "work independently … without requiring per-controller configuration." |
| 2  | Thin wrappers, never replacements | `data/scope.go` is 1–3-line delegations to sqlx; builder takes raw `*cobra.Command`; fragment takes raw `chi.Router`. `fxlog` **re-exports** `slog.Any/String/…` so callers need one import — re-surfaces the primitive, never hides it. |
| 3  | Decentralized config declarations | `config/var.go:24` doc, verbatim: "Var allows each part of the application (and fx itself) to declare their own configuration variables … without requiring changes to the config package directly." `NewVar` self-registers. |
| 4  | Context as carry-bag              | `config/context.go`, `data/context.go`, `data/scope.go` — the transaction itself rides `ctx` (`getTx`/`setTx`), enabling savepoint-free nested transactions.                                                                       |
| 5  | Everything optional               | Dozens of "Allow X" commits; `NewSource(nil)` → default `EnvProvider`; conditional defaults applied only when unset.                                                                                                                   |
| 6  | Operational concerns first-class  | `print-config` ("to enable scripting"), `data psql`, the migration recover/resync/sync saga — and the worker's choice of a visible CAS over `FOR UPDATE` locks (see G1). Visibility is a first-class design input.                     |
| 7  | Punt distributed systems to infra | **The worker (see G1) is the canonical instance** — distribution punted to Postgres CAS + OS process scheduling, deliberately, after rejecting an in-process parallel design.                                                          |
| 8  | Convention over options, one deep | `config/source.go:50` teaches how to "roll your own `Configure()`"; `14c42b7` deliberately *removed* baked-in default commands to make them opt-in.                                                                                    |

## The gaps — philosophies he writes by but never wrote down

### G1. Subtraction via stronger primitives — the win is the deletion

The most load-bearing principle the eight don't name, and the one that *grounds* #7. Per
the maintainer: he spent significant effort on a distributed-worker design — in-process
parallelism, spawning e.g. 3 workers per process — then rejected it for **stronger
primitives already on hand**: OS process scheduling for parallelism, Postgres transactions
for coordination. Going that route "removed and simplified the worker code a ton." The
code confirms it:

- `worker/worker.go` runs a **single** polling goroutine with a `sync.Mutex` serializing
  `workOnce`. No goroutine pool. Scale-out is *more processes*, not more threads — the OS
  is the scheduler.
- `worker/jobs.go` coordinates competing worker processes through a Postgres
  compare-and-swap: `SELECT` a pending job, then `UPDATE … WHERE id=$ AND status='pending'
  RETURNING *`; only one worker wins the transition. `ORDER BY RANDOM()` minimizes
  collisions under load.
- And it rejects an even *stronger* lock primitive for an operational reason — verbatim
  comment: *"we could use FOR UPDATE locks but this means the 'processing' status update
  won't be visible to other workers and we lose visibility into jobs that are actually
  under processing."* So the primitive choice is subordinated to #6 (visibility).

The meta-heuristic: when two designs compete, prefer the one that leans on an existing
strong primitive (OS, Postgres) **and leaves you with less code** — net deletion is the
signal you chose right. This is the *authentic* subtraction discipline — deleting his own
complexity. The eight imply minimalism but never state "the simplification is the proof".

### G2. Provenance + speculation discipline — design in prose, build on demand

Two layers the eight miss:

- **Capabilities originate from real downstream apps**, not greenfield design. The whole
  framework was extracted via `git subtree`; features arrive as "Import OpenAI client from
  x9". Abstractions are earned by use, not posited.
- **Speculation lives in comments; implementation defers to need.** At birth,
  `config/source.go` carried a 9-line doc comment pre-declaring "could be changed to
  provide values from etcd or hashicorp vault" and left a real seam — a `src *Source`
  parameter **accepted but ignored**. The `config.Provider` interface that fulfils that
  exact speculation landed **2.5 years later** (`8ad5c6d`), only when a non-env source was
  actually needed. Design intent is narrated cheaply in prose and stubbed with cheap seams
  (ignored params, marker interfaces, `// TODO` *inside* the interface); the heavyweight
  build waits.

This is **not** dogmatic YAGNI. The migration `Source` shipped as a 5-member family
(`FromDir/FromFS/FromConfig/FromAuto/Embed`) on day one — built broad and eager — because
each member is a free one-line closure over existing load logic. The discriminator is
cost/uncertainty: build breadth when it's a cheap closure family; defer when it's a weighty
interface of unknown shape. Abstraction *timing* is a judgment call, not a law.

### G3. Ship rough, harden under use

Labeled-rough first cuts are normal and acceptable: "very basic mailer module", "Basic
pagination implementation", "Adds basic DOCS, for now", "Initial config.Provider
abstraction", and "Initial worker re-implementation" (a v2 — the first worker was
scrapped). Each is later hardened where reality pushed back: "More robust handling of
TTY-interactivity", "Make ctrlc a bit more robust", config "try harder to find .env",
worker "process jobs a bit faster". The eight read as steady-state design tests and are
silent on this maturation process.

### G4. Caller-safety over author-convenience (the nil-interface stance)

`7cb6ed1` rewrote validators from `*FieldError` to plain `error`, with the commit body:
*"Returning `*FieldError` and letting the client code box it into error interface leads to
`if err != nil` check breaking unexpectedly all over the place. Go language really should
get rid of this semantic."* He gives up compile-time `*FieldError` typing — and adds a
*runtime* `panic` guard in `Multi` — specifically so a caller's `if err != nil` never lies
(the typed-nil footgun). The error system is duck-typed for the same reason: decoration
reads metadata via anonymous-interface assertions (`err.(interface{ Code() string })`), so
anything can opt into being decorated without importing a named interface. Consumer
ergonomics outrank the author's static-typing convenience.

### G5. Bugs are contract smells; honesty over closure

The connection-leak fix (`7bd4f3f`) was a literal two-line hoist (move `newDataContext`
out of the middleware closure so one pool is shared). But the commit body doesn't declare
victory — it says *"We have to review the fragment code again, to make sure that didn't
happen or that the contract is clear."* A bug is treated as evidence of an unclear seam,
and the residual uncertainty is left visible rather than papered over.

### G6. Anti-dogmatism — conventions are judgment-applied heuristics

"Replace `log` with `fxlog` **where make sense**" — an explicit refusal to apply his own
migration blanket. Same judgment shows up in G2's build-broad-vs-defer split and G1's
reject-the-stronger-lock-for-visibility call. The stated philosophy's only nod is its
closing line ("Minimalism here is a discipline … not a goal in itself"), scoped narrowly
to minimalism. The broader meta-rule — *the principles are heuristics applied with
judgment, not laws* — is under-stated.

## Operating rhythm (the generative context)

Development is **episodic and burst-driven**: long dormancy punctuated by intense
single-day sessions (the 2023-06-21 birth; a ~20-commit burst on 2023-12-28/29;
2024-02-25/26; 2024-07-11; 2025-07-28; 2026-01-11), with multi-month gaps between. This
rhythm *generates* G2's provenance discipline — FX grows when a downstream app forces a
need, then sits. It also explains why speculation is parked in prose: between bursts the
comment is the only carrier of intent.

## House-style fingerprint (code-craft, not philosophy — spine only)

- Scoped `if x := …; cond { … } else { … }` even on happy-path returns.
- Named-return values + `defer` to decorate the error (`func newScope() (…, err error)`;
  `defer errutil.Wrap("worker", &err)`).
- Layer-prefix error wrapping (`tx: %w`, `migrator: %w`, `worker:`) + sentinel errors with
  `IsX` predicates (`ErrEmpty`/`IsEmpty`, `ErrNoMigrations`/`IsNoMigrations`).
- Compile-time interface assertions `var _ Iface = impl{}`; interface + lowercase impl.
- Two-tier accessors: a terse happy-path one that assumes correct wiring
  (`data.FromContext` panics if absent) plus an explicit safe variant
  (`LookupFromContext`, `config.GetOK`).
- Three-state `(value, ok, error)` returns to separate absent / error / empty.
- Re-exporting the wrapped primitive's API (`fxlog` aliasing `slog.*`).
- Graceful degradation: non-fatal config errors are logged, never block startup.
- Evolution is layer-bottom-up and usually non-breaking via parallel constructors (`New`
  preserved, `NewWithFragments` added) — except where a commit flags "Break API a bit".

## What to consider promoting into `philosophy.md`

Proposed, not decided — the call is the maintainer's:

1. **G1 (subtraction via stronger primitives)** is the biggest omission and the sharpest
   lever: it both grounds #7 and supplies the missing "simplification is the proof"
   heuristic. The worker decision is its worked example.
2. **G2 (provenance + prose-speculation, deferred build)** is the deepest meta-rule and
   the most useful to state for contributors deciding *when* to add an abstraction.
3. **G3, G4, G6** are real but lower-altitude; could fold into existing principles (G3 → a
   note on iterative maturation, G4 → under #2, G6 → expand the closing line).
