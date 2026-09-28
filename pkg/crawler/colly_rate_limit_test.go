/*
==========
Cariddi
==========

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see http://www.gnu.org/licenses/.

	@Repository:  https://github.com/edoardottt/cariddi

	@Author:      edoardottt, https://edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package crawler_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/edoardottt/cariddi/pkg/crawler"
	"github.com/edoardottt/cariddi/pkg/input"
	"github.com/projectdiscovery/ratelimit"
)

const testPage = `<html><head><title>cariddi</title></head><body>cariddi rate limiter test page</body></html>`

// newTestServer starts a local HTTP server that records the arrival
// time of every request it receives.
func newTestServer(t *testing.T) (*httptest.Server, *[]time.Time, *sync.Mutex) {
	t.Helper()

	mu := &sync.Mutex{}
	timestamps := &[]time.Time{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		*timestamps = append(*timestamps, time.Now())
		mu.Unlock()

		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, testPage)
	}))

	t.Cleanup(srv.Close)

	return srv, timestamps, mu
}

// newTestScan builds a minimal crawler.Scan hitting the given target.
// Concurrency is 1 so requests are serialized and timing is predictable.
func newTestScan(target string) *crawler.Scan {
	return &crawler.Scan{
		Target:      target,
		Concurrency: 1,
		Timeout:     input.TimeoutRequest,
	}
}

// requestGap returns the duration between the first and last recorded
// request under the given lock.
func requestGap(timestamps *[]time.Time, mu *sync.Mutex) time.Duration {
	mu.Lock()
	defer mu.Unlock()

	first := (*timestamps)[0]
	last := (*timestamps)[len(*timestamps)-1]

	return last.Sub(first)
}

func TestCrawlerRateLimitThrottlesRequests(t *testing.T) {
	srv, timestamps, mu := newTestServer(t)

	// Limit the whole scan to 1 request per second.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	limiter := ratelimit.New(ctx, 1, time.Second)

	start := time.Now()
	crawler.New(newTestScan(srv.URL), limiter)
	elapsed := time.Since(start)

	mu.Lock()
	requests := len(*timestamps)
	mu.Unlock()

	if requests < 2 {
		t.Fatalf("expected at least 2 requests, got %d", requests)
	}

	// The second request must wait for the next token (~1 second),
	// so consecutive requests must be clearly spaced apart.
	if gap := requestGap(timestamps, mu); gap < 800*time.Millisecond {
		t.Errorf("first/last request gap = %v, want >= 800ms for 1 rps (elapsed %v)", gap, elapsed)
	}
}

func TestCrawlerUnlimitedHasNoThrottling(t *testing.T) {
	srv, timestamps, mu := newTestServer(t)

	limiter := ratelimit.NewUnlimited(context.Background())

	start := time.Now()
	crawler.New(newTestScan(srv.URL), limiter)
	elapsed := time.Since(start)

	mu.Lock()
	requests := len(*timestamps)
	mu.Unlock()

	if requests < 2 {
		t.Fatalf("expected at least 2 requests, got %d", requests)
	}

	// Without rate limiting the requests are only serialized by the
	// crawler itself and must complete well under one second.
	gap := requestGap(timestamps, mu)
	if gap >= time.Second {
		t.Errorf("unlimited limiter throttled requests: gap = %v, elapsed %v", gap, elapsed)
	}
}
