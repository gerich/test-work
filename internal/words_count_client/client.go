package words_count_client

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// NewClient
func NewClient(client http.Client, ctx context.Context, word string) *WordCountClientImpl {
	return &WordCountClientImpl{client, ctx, word}
}

type WordCountClient interface {
	ProcessURL(url.URL) (int, error)
}

// WordCountClient
type WordCountClientImpl struct {
	client http.Client
	ctx    context.Context
	word   string
}

// ProcessURL
func (c *WordCountClientImpl) ProcessURL(urlForParse url.URL) (int, error) {
	req := &http.Request{
		Method: http.MethodGet,
		URL:    &urlForParse,
	}
	req.WithContext(c.ctx)
	resp, err := c.client.Do(req)

	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, nil
	}

	return strings.Count(string(body), c.word), nil
}
