package coda

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"fx.prodigy9.co/config"
)

const APIPrefix = "https://coda.io/apis/v1"

var CodaTokenConfig = config.Str("CODA_TOKEN")

type Client struct {
	token string
	http  http.Client
}

func NewClient(cfg *config.Source) *Client {
	return &Client{
		token: config.Get(cfg, CodaTokenConfig),
		http:  http.Client{},
	}
}

func (c *Client) ListDocs() (*More[*Doc], error) {
	more := &More[*Doc]{}
	if err := c.CallAPI(more, nil, "GET", "/docs", nil); err != nil {
		return nil, err
	}
	return more, nil
}
func (c *Client) ListPages(docID string) (*More[*Page], error) {
	more, p := &More[*Page]{}, "/docs/"+url.PathEscape(docID)+"/pages"
	if err := c.CallAPI(more, nil, "GET", p, nil); err != nil {
		return nil, err
	}
	return more, nil
}
func (c *Client) ListTables(docID string) (*More[*Table], error) {
	more, p := &More[*Table]{}, "/docs/"+url.PathEscape(docID)+"/tables?tableType=table"
	if err := c.CallAPI(more, nil, "GET", p, nil); err != nil {
		return nil, err
	} else {
		return more, nil
	}
}
func (c *Client) ListColumns(docID, tableID string) (*More[*Column], error) {
	more, p := &More[*Column]{}, "/docs/"+url.PathEscape(docID)+"/tables/"+url.PathEscape(tableID)+"/columns"
	if err := c.CallAPI(more, nil, "GET", p, nil); err != nil {
		return nil, err
	} else {
		return more, nil
	}
}
func (c *Client) ListRows(docID, tableID string) (*More[*Row], error) {
	return c.ListRowsWithQuery(docID, tableID, nil)
}
func (c *Client) ListRowsWithQuery(docID, tableID string, query map[string]string) (*More[*Row], error) {
	qs := url.Values{}
	if len(query) > 0 {
		for key, value := range query {
			qs.Add("query", fmt.Sprintf(`%s="%s"`, c.escapeQuery(key), c.escapeQuery(value)))
		}
	}

	more, p := &More[*Row]{}, "/docs/"+url.PathEscape(docID)+"/tables/"+url.PathEscape(tableID)+"/rows"
	if err := c.CallAPI(more, nil, "GET", p, qs); err != nil {
		return nil, err
	} else {
		return more, nil
	}
}

func (c *Client) CallAPI(result, payload any, method, path string, qs url.Values) error {
	u, err := url.Parse(APIPrefix + path + "?valueFormat=rich")
	if err != nil {
		return fmt.Errorf("coda: %w", err)
	}

	if len(qs) > 0 {
		u.RawQuery = qs.Encode()
	}

	buffer := &bytes.Buffer{}
	if payload != nil {
		if err := json.NewEncoder(buffer).Encode(payload); err != nil {
			return fmt.Errorf("coda: %w", err)
		}
	}

	req, err := http.NewRequest(method, u.String(), buffer)
	if err != nil {
		return fmt.Errorf("coda: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.http.Do(req)
	defer resp.Body.Close()

	switch {
	case err != nil:
		return fmt.Errorf("coda: %w", err)
	case 300 <= resp.StatusCode && resp.StatusCode < 400:
		return fmt.Errorf("coda: redirect required")
	case 400 <= resp.StatusCode && resp.StatusCode < 600:
		if buf, err := io.ReadAll(resp.Body); err != nil {
			return fmt.Errorf("coda: %d and failed to read response: %w", resp.StatusCode, err)
		} else {
			return fmt.Errorf("coda: %d: %s", resp.StatusCode, string(buf))
		}
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("coda: failed to read response: %w", err)
		}
	}
	return nil
}

func (c *Client) escapeQuery(s string) string {
	s = strings.Replace(s, `"`, `\"`, -1)
	s = strings.Replace(s, `=`, `\=`, -1)
	return `"` + s + `"`
}
