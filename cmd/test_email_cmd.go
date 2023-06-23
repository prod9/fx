package cmd

import (
	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/errutil"
	"fx.prodigy9.co/mailer"

	"github.com/spf13/cobra"
)

var testEmailCmd = &cobra.Command{
	Use:   "test-email",
	Short: "Sends a test email to check if SMTP configuration works",
	RunE:  runTestEmailCmd,
}

func runTestEmailCmd(cmd *cobra.Command, args []string) (err error) {
	defer errutil.Wrap("test-email", &err)

	cfg := config.Configure()
	recipient := prompts.New(cfg, args).Str("email recipient")
	mail := &mailer.Mail{
		From:     "testdev@prodigy9.co",
		To:       []string{recipient},
		Subject:  "Test Email",
		HTMLBody: "<strong>Test Email</strong>",
		TextBody: "Test Email",
	}

	return mailer.Send(cfg, mail)
}
