# Send-a-file example

A minimal "send a file to a friend" service built on FX's built-in `files` app and the
local `blobserver`. It shows the whole split: **object bytes live on disk** (via
`blobserver`), **metadata lives in Postgres** (via `files`), and **presigned URLs** tie
them together — the browser/CLI transfers bytes directly, the app only mints URLs.

A *drop* is a public share addressed by an unguessable token. Create a drop, upload one
file to it, hand the link to a friend, they download it.

## How it fits together

* `drops/` — a fragment with a `drops` table and `POST /drops` (mints a random token).
  It mounts FX's `files.Controller` under `/d/{token}/file`, with
  `files.WithMode(files.ModeReadWrite)` (the controller defaults to read-only, so
  writes are opt-in) and `files.WithOwnerIDFunc` to resolve the drop from the URL token.
* `files.App` — FX's built-in fragment (`app/files`); owns the `files` metadata table,
  the presigned upload/download endpoints, and a cleanup worker. Mounted, not modified.
* `blobserver` — serves the object bytes from local disk; `STORAGE_URL`'s `http://`
  scheme points `blobstore` at it (see `.env`).

## Cleanup worker

`files.App` registers a `files.cleanup` worker that reconciles the store against the
`files` table — deleting objects with no owning row and pruning rows whose upload never
arrived (older than 24h).

> The `files.cleanup` sweep deletes every bucket object without an owning `files` row, so
> `files.App` must own its `STORAGE_URL` bucket exclusively.

## Routes

| Method | Path         | Purpose                                             |
|--------|--------------|-----------------------------------------------------|
| POST   | `/drops`     | Create a drop → `{ "token": "…" }`                  |
| POST   | `/d/{token}` | Register a file → `{ file_info, upload_url }`       |
| PUT    | *upload_url* | Upload bytes directly to `blobserver` (presigned)   |
| GET    | `/d/{token}` | 307 → presigned GET; the share link a friend opens  |
| DELETE | `/d/{token}` | Remove the file                                     |

## Running

Two servers: the API and the local `blobserver`. Both read `.env`.

```sh
cd examples/sendfile
createdb sendfile
```

**Migrate.** `files.App`'s migration is embedded (it lives in `app/files/`, not under
this directory), and `migrator.LoadAuto` scans CWD for `*.sql` and short-circuits on the
local `drops` migration before reaching the embedded set. So run migrations from a clean
directory to force the embedded path (this is a known footgun — see `docs/TODO.md`
"`migrator.LoadAuto`: merge disk + embed"):

```sh
export DATABASE_URL=postgres:///sendfile?sslmode=disable
export ALWAYS_YES=1
go build -o /tmp/sendfile .
( cd /tmp && /tmp/sendfile data migrate )   # applies create_files + create_drops
```

**Serve** (two terminals, both from `examples/sendfile` so `.env` loads):

```sh
go run . store serve   # terminal 1 — blobserver on :9500, bytes under tmp/blobstore/
go run . serve         # terminal 2 — API on :3000
```

## Send a file

```sh
# 1. create a drop
TOKEN=$(curl -s -XPOST localhost:3000/drops | sed -n 's/.*"token":"\([a-f0-9]*\)".*/\1/p')

# 2. register the file, get a presigned upload URL
echo "secret weekend plans" > plans.txt
SIZE=$(wc -c < plans.txt | tr -d ' ')
UPLOAD_URL=$(curl -s -XPOST "localhost:3000/d/$TOKEN" \
  -H 'content-type: application/json' \
  -d "{\"original_name\":\"plans.txt\",\"content_type\":\"text/plain\",\"content_length\":$SIZE}" \
  | sed -n 's/.*"upload_url":"\([^"]*\)".*/\1/p')

# 3. upload the bytes straight to blobserver
curl -XPUT --upload-file plans.txt "$UPLOAD_URL"

# 4. the share link — what you send your friend
curl -sL "localhost:3000/d/$TOKEN"     # → secret weekend plans
```

The uploaded bytes land at `tmp/blobstore/sendfile/token/drop/<dropID>/<fileID>`;
the row in `files` records the name, content type, and size.
