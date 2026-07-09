# Blob Storage

**Status:** implemented

The `blobstore` package is an S3-compatible object client. It hands out presigned
URLs so clients (browsers, the `files/` app) transfer bytes directly against the
storage endpoint — the app mints URLs and never proxies the payload.

```go
getURL, err := blobstore.PresignedGetURL(ctx, "uploads/photo.jpg")
putURL, err := blobstore.PresignedPutURL(ctx, "uploads/photo.jpg",
	blobstore.WithContentType("image/jpeg"))
err = blobstore.DeleteObject(ctx, "uploads/photo.jpg")
```

## Configuration

* `STORAGE_URL` — endpoint, credentials, and bucket in one URL. The scheme selects
  the transport:
  * `s3://key:secret@endpoint/bucket` — TLS (the live default).
  * `https://key:secret@endpoint/bucket` — TLS, explicit.
  * `http://key:secret@endpoint/bucket` — plaintext, for the local `blobserver`.

A plaintext (`http`) endpoint is treated as the local `blobserver`: the client pins a
region so presigning skips the `GetBucketLocation` handshake that a live S3 performs.

## Local development — `blobserver`

`blobserver` (`blobstore/blobserver`) is a minimal, real object store over a local
directory. It serves the verbs `blobstore` emits against S3 — GET/PUT/DELETE plus a
`ListObjectsV2` for reconciliation — persisting objects as plain files, with no auth or
ACLs. It exists so you can develop against on-disk blobs before pointing `STORAGE_URL` at
a live S3.

```sh
go run . store serve   # serve BLOBSERVER_DIR over BLOBSERVER_ADDR
```

Point the client at it:

```sh
STORAGE_URL=http://key:secret@localhost:9500/mybucket
```

Presigned URLs then resolve to `blobserver`, and the `files/` upload/download flow
works unchanged.

* `BLOBSERVER_ADDR` — listen address (default `0.0.0.0:9500`).
* `BLOBSERVER_DIR` — storage directory (default `tmp/blobstore`, git-ignored).

## `store` commands

The `store` command group (added by `AddDefaults`) operates against whatever
`STORAGE_URL` points at — live S3 or the local `blobserver`:

| Command                      | Purpose                                          |
|------------------------------|--------------------------------------------------|
| `store serve`                | Run the local `blobserver`.                      |
| `store upload <key> [src]`   | Upload a file (or stdin) via a presigned PUT.    |
| `store download <key> [dest]`| Download to a file (or stdout) via a presigned GET. |
| `store presign-get <key>`    | Print a presigned GET URL.                       |
| `store presign-put <key>`    | Print a presigned PUT URL.                       |
| `store delete <key>`         | Delete an object.                                |

`upload`/`download` default the source/destination to stdin/stdout, so they pipe:

```sh
cat photo.jpg | go run . store upload uploads/photo.jpg
go run . store download uploads/photo.jpg > photo.jpg
```
