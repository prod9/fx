# Error Utilities

**Status:** accepted

The `errutil` package provides helpers for decorating and collecting errors.

* `errutil.Wrap(name, &err)` — Intended for use in a `defer`. Prefixes the error with
  `name` if `*err` is non-nil.
* `errutil.WithCode(err, code)` — Attaches a string error code to the error (useful
  for API error responses).
* `errutil.WithData(err, data)` — Attaches arbitrary context data to the error.
* `errutil.NewCoded(code, msg, data)` — Creates a new error with code, message, and
  data.
* `errutil.Decorate(err)` — Wraps an error in a `decoratedErr` for JSON serialization.
* `errutil.Aggregate[T](slice, func)` — Runs a function on each element in parallel,
  collecting all errors into a single aggregated error.
* `errutil.AggregateWithTags[T](slice, func)` — Like `Aggregate` but each error is
  tagged with a label for identification.
