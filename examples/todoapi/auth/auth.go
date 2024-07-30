package auth

import "fx.prodigy9.co/app"

var App = app.Build().
	Description("Basic username/password authentication").
	Commands(
		NewUserCmd,
	).
	Controllers(
		UserCtr{},
		SessionCtr{},
	)
