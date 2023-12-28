package auth

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"fx.prodigy9.co/cmd/prompts"
	"fx.prodigy9.co/config"
	"fx.prodigy9.co/data"
	"github.com/spf13/cobra"
)

var NewUserCmd = &cobra.Command{
	Use:   "new-user [username]",
	Short: "Creates a new user",
	Run:   runNewUserCmd,
}

func runNewUserCmd(cmd *cobra.Command, args []string) {
	cfg := config.Configure()
	p := prompts.New(cfg, args)
	action := &CreateUser{
		Username:             p.Str("username"),
		Password:             p.SensitiveStr("password"),
		PasswordConfirmation: p.SensitiveStr("confirm password again"),
	}

	if err := action.Validate(); err != nil {
		log.Fatalln(err)
	}

	db, err := data.Connect(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	ctx := data.NewContext(context.Background(), db)
	user := &User{}
	if err := action.Execute(ctx, user); err != nil {
		log.Fatalln(err)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(user); err != nil {
		log.Fatalln(err)
	}
}
