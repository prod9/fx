package mailer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"fx.prodigy9.co/config"
	"fx.prodigy9.co/errutil"
)

var PostmarkTokenConfig = config.Str("POSTMARK_TOKEN")

type Mail struct {
	From string
	To   []string
	CC   []string

	Subject  string
	HTMLBody string
	TextBody string
}

const mailFormat = "From: %s\r\nCc: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s"

func Send(cfg *config.Source, mail *Mail) (err error) {
	defer errutil.Wrap("mailer", &err)

	payload := map[string]string{}
	if mail.From != "" {
		payload["From"] = mail.From
	}
	if len(mail.To) > 0 {
		payload["To"] = strings.Join(mail.To, ", ")
	}
	if len(mail.CC) > 0 {
		payload["Cc"] = strings.Join(mail.CC, ", ")
	}
	if mail.Subject != "" {
		payload["Subject"] = mail.Subject
	}
	if mail.HTMLBody != "" {
		payload["HtmlBody"] = mail.HTMLBody
	}
	if mail.TextBody != "" {
		payload["TextBody"] = mail.TextBody
	}

	buf, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.postmarkapp.com/email", strings.NewReader(string(buf)))
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Postmark-Server-Token", config.Get(cfg, PostmarkTokenConfig))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	buf, err = ioutil.ReadAll(resp.Body)
	switch {
	case err != nil:
		return fmt.Errorf("postmark: status %d: failed to read response: %w", resp.StatusCode, err)
	case !(200 <= resp.StatusCode && resp.StatusCode < 300):
		return fmt.Errorf("postmark: status %d: %s", resp.StatusCode, string(buf))
	}

	log.Println("postmark:", string(buf))
	return nil
}
