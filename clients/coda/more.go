package coda

import "errors"

var ErrNoMore = errors.New("coda: no more pages")

/*
	{
	  "items": [],
	  "href": "https://coda.io/apis/v1/docs?limit=20",
	  "nextPageToken": "eyJsaW1pd",
	  "nextPageLink": "https://coda.io/apis/v1/docs?pageToken=eyJsaW1pd"
	}
*/
type More[T any] struct {
	Items         []T    `json:"items"`
	NextPageToken string `json:"nextPageToken"`
	NextPageLink  string `json:"nextPageLink"`
}

func LoadMore[T any](c *Client, more *More[T]) (*More[T], error) {
	if more.NextPageLink == "" {
		return nil, ErrNoMore
	}

	m := &More[T]{}
	if err := c.CallAPI(m, nil, "GET", m.NextPageLink, nil); err != nil {
		return nil, err
	} else {
		return m, nil
	}
}
