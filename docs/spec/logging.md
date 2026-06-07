# Logging

**Status:** accepted

Logging in fx is done using the `fxlog` subpackage. It is pre-configured to output
pretty structured logs by default. It has 3 basic functions, mirroring the standard
library log package:

* `fxlog.Log` — Logs general messages.
* `fxlog.Error` — Logs errors.
* `fxlog.Fatal` — Logs fatal errors and exits the application.

For `Error` and `Fatal`, there are also `Errorf` and `Fatalf` variants that simply call
`fmt.Errorf` to format the message for you.

Log output is set to the default `zerolog` logger by default. There are a few ways to
override the output and customize logging behavior:

1. Switch the sink: set `LOG_SINK=slog` to redirect FX log outputs to `log/slog`'s
   default logger and configure `log/slog` normally as you would in any other Go
   application.

2. Set a custom `log/slog` or `zerolog` logger (initialized outside of fx) by using
   `SetSink`:

   ```go
   // zerolog
   zl := zerolog.New()
   fxlog.SetSink(fxlog.NewZerlogSink(zl))

   // slog
   sl := slog.New(slog.NewTextHandler(os.Stderr, nil))
   fxlog.SetSink(fxlog.NewSlogSink(sl))
   ```

3. Create your own `fxlog.Sink` implementation for maximum customization or if you
   wish to use a different logging library that's not provided out of the box:

   ```go
   mysink := NewCustomSink()
   fxlog.SetSink(mysink)
   ```
