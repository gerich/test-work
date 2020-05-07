package words_count_client

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

const word = "Go"

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

//NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func newTestClient(fn RoundTripFunc) http.Client {
	return http.Client{
		Transport: RoundTripFunc(fn),
	}
}
func Test_WordCountClient_ProcessUrl(t *testing.T) {
	expectedURL := "https://foo.com/bar?baz=asd"
	expectedCount := 3

	httpClient := newTestClient(func(req *http.Request) *http.Response {
		if req.URL.String() != expectedURL {
			t.Error("Урлы не совпали", expectedURL, req.URL.String())
		}

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(`go Go asd13Goasd Go go`)),
			Header:     make(http.Header),
		}
	})

	client := NewClient(httpClient, context.Background(), word)
	preparedURL, _ := url.Parse(expectedURL)
	actualCount, err := client.ProcessURL(*preparedURL)
	if err != nil {
		t.Error("Ошибка при успешном кейсе", err)
	}

	if expectedCount != actualCount {
		t.Error("Колво слов не совпало с ожидаемым", expectedCount, actualCount)
	}
}
