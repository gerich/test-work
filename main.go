package main

import (
	"bufio"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"time"
	wcc "wildberries-test/internal/words_count_client"
)

const (
	word                     = "Go"
	maxParallelRequestsCount = 5
	sleepTimeMs              = 10
)

type result struct {
	err   error
	url   string
	count int
}

func scheduler(ctx context.Context, in chan url.URL, out chan<- result, client wcc.WordCountClient) {
	var currentRequestsCount int32 = 0

	for {
		for atomic.LoadInt32(&currentRequestsCount) >= maxParallelRequestsCount {
			time.Sleep(sleepTimeMs * time.Millisecond)
		}
		select {
		case nextUrl := <-in:
			go func(counter *int32) {
				count, err := client.ProcessURL(nextUrl)
				out <- result{err, nextUrl.String(), count}
				atomic.AddInt32(counter, -1)
			}(&currentRequestsCount)
			atomic.AddInt32(&currentRequestsCount, 1)
		case <-ctx.Done():
			close(out)
			return
		}
	}
}

func processResult(ctx context.Context, resultChan <-chan result, logger *log.Logger) {
	for {
		select {
		case res := <-resultChan:
			logger.Println(res)
		case <-ctx.Done():
			return
		}
	}
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	in := make(chan url.URL, maxParallelRequestsCount)
	out := make(chan result, maxParallelRequestsCount)
	client := wcc.NewClient(http.Client{Timeout: 5 * time.Second}, ctx, word)
	logger := log.New(os.Stderr, "", 0)
	go processResult(ctx, out, logger)
	go scheduler(ctx, in, out, client)
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		if text == "\n" {
			return
		}
		parsedUrl, err := url.Parse(strings.Replace(text, "\n", "", 1))
		if err != nil {
			cancel()
			return
		}
		in <- *parsedUrl
	}
}
