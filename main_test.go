package main

import (
	"context"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

type clientMock struct {
	countMap map[string]int
}

func (c *clientMock) ProcessURL(u url.URL) (int, error) {
	count, found := c.countMap[u.String()]
	if !found {
		count = -1
	}

	return count, nil
}

func TestScheduler_OneRequest_Success(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resultChan := make(chan result, maxParallelRequestsCount)
	inChan := make(chan url.URL, maxParallelRequestsCount)
	expected := map[string]int{
		"https://foo.com/bar?count=1": 1,
		"https://foo.com/bar?count=2": 2,
		"https://foo.com/bar?count=3": 3,
		"https://foo.com/bar?count=4": 4,
		"https://foo.com/bar?count=5": 5,
		"https://foo.com/bar?count=6": 6,
		"https://foo.com/bar?count=7": 7,
		"https://foo.com/bar?count=8": 8,
	}
	expectedRequestsCount := len(expected)
	client := &clientMock{countMap: expected}

	go scheduler(ctx, inChan, resultChan, client)

	actualRequestsCount := 0
	wg := sync.WaitGroup{}
	wg.Add(2)
	timeout := time.After(sleepTimeMs * time.Millisecond)
	go func() {
		for expectedURL := range expected {
			parsedURL, _ := url.Parse(expectedURL)
			inChan <- *parsedURL
		}
		wg.Done()
	}()
	go func() {
		for {
			select {
			case res := <-resultChan:
				actualRequestsCount++
				assertResult(t, res, expected)
				timeout = time.After(2 * sleepTimeMs * time.Millisecond + 1)
			case <-timeout:
				wg.Done()
				return
			}
		}
	}()
	wg.Wait()
	if actualRequestsCount != expectedRequestsCount {
		t.Error("Колво обработанных запросов не совпало", expectedRequestsCount, actualRequestsCount)
	}
}

func assertResult(t *testing.T, res result, expectedMap map[string]int) {
	expectedCount, found := expectedMap[res.url]
	if !found {
		t.Error("Колво для урла не установленно", res.url)
	}

	if res.err != nil {
		t.Error("Ошибка во время обработки", res.err)
	}

	if res.count != expectedCount {
		t.Error("Колво не совпало", expectedCount, res.count)
	}
}

type mockWriter struct {
	result string
	sync.Mutex
}

func (m *mockWriter) Write(p []byte) (int, error) {
	m.Lock()
	m.result = string(p)
	m.Unlock()

	return len(p), nil
}

func TestProcessResult(t *testing.T) {
	mock := &mockWriter{"", sync.Mutex{}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resultChan := make(chan result)
	logger := log.New(mock, "", 0)

	go processResult(ctx, resultChan, logger)

	expectedURL := "https://foo.com/bar"
	expectedCount := 12
	resultChan <- result{nil, expectedURL, expectedCount}
	time.Sleep(time.Millisecond)
	mock.Lock()
	if !strings.Contains(mock.result, expectedURL) {
		t.Error("URL не найден", expectedURL, mock.result)
	}

	if !strings.Contains(mock.result, strconv.Itoa(expectedCount)) {
		t.Error("Колво неуказанно", expectedCount, mock.result)
	}
	mock.Unlock()
}
