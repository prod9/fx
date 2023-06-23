package openai

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"fx.prodigy9.co/config"
)

var (
	KeyConfig = config.Str("OPENAI_KEY")
	OrgConfig = config.Str("OPENAI_ORG")
)

const (
	Model     = "gpt-3.5-turbo"
	MaxTokens = 512
)

type Client struct {
	cfg    *config.Source
	sysMsg string
}

func New(cfg *config.Source, sysMsg string) *Client {
	return &Client{cfg, sysMsg}
}

func (c *Client) CompleteText(text string) (string, error) {
	data := map[string]any{
		"model": Model,
		"messages": []map[string]any{
			{"role": "system", "content": c.sysMsg},
			{"role": "user", "content": text},
		},
		"temperature": 0.6,
		"n":           1,
		"max_tokens":  MaxTokens,
	}

	payload := &bytes.Buffer{}
	if err := json.NewEncoder(payload).Encode(data); err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", payload)
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+config.Get(c.cfg, KeyConfig))
	req.Header.Set("OpenAI-Organization", config.Get(c.cfg, OrgConfig))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return "", err
	}

	rawBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("http request failed: %w", err)
	}
	if !(200 <= resp.StatusCode && resp.StatusCode < 300) {
		return "", errors.New("status " + strconv.Itoa(resp.StatusCode) + ": " + string(rawBody))
	}

	/* {
	  "id": "chatcmpl-123",
	  "object": "chat.completion",
	  "created": 1677652288,
	  "choices": [{
	    "index": 0,
	    "message": {
	      "role": "assistant",
	      "content": "\n\nHello there, how may I assist you today?",
	    },
	    "finish_reason": "stop"
	  }],
	  "usage": {
	    "prompt_tokens": 9,
	    "completion_tokens": 12,
	    "total_tokens": 21
	  }
	}*/
	body := &struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}{}

	if err := json.NewDecoder(bytes.NewReader(rawBody)).Decode(body); err != nil {
		return "", fmt.Errorf("status "+strconv.Itoa(resp.StatusCode)+": error decoding response: %w", err)
	}

	return strings.TrimSpace(body.Choices[0].Message.Content), nil
}
