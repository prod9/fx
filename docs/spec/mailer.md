# Mailer

**Status:** accepted

The `mailer` package sends transactional emails via Postmark.

```go
err := mailer.Send(cfg, &mailer.Mail{
	From:     "noreply@example.com",
	To:       []string{"user@example.com"},
	Subject:  "Welcome!",
	HTMLBody: "<h1>Hello</h1>",
	TextBody: "Hello",
})
```

## Configuration

* `POSTMARK_TOKEN` — Postmark server API token.
