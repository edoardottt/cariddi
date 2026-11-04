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
	"bytes"
	"fmt"
	"io"
	"log"
	"maps"
	"net"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/edoardottt/cariddi/v2/pkg/crawler"
	"github.com/edoardottt/cariddi/v2/pkg/input"
)

// newTestProxy returns a proxy for the http URLs, counting the requests it
// forwards: they reach the target with the header Via.
func newTestProxy(t *testing.T, requests *atomic.Int32) *httptest.Server {
	t.Helper()

	return serve(t, proxyHandler(t, requests))
}

// proxyHandler is the handler of the proxy returned by newTestProxy.
func proxyHandler(t *testing.T, requests *atomic.Int32) http.HandlerFunc {
	t.Helper()

	transport := &http.Transport{}
	t.Cleanup(transport.CloseIdleConnections)

	return func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		out := r.Clone(r.Context())
		out.RequestURI = ""
		out.Header.Set("Via", "test proxy")

		resp, err := transport.RoundTrip(out)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)

			return
		}
		defer resp.Body.Close()

		maps.Copy(w.Header(), resp.Header)
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

// TestNewProxies checks that with one or more proxies every request (the
// sitemap probe and the redirects too) uses a proxy, in turn, and that the
// results are the ones of the crawl without proxies.
func TestNewProxies(t *testing.T) {
	var total, direct atomic.Int32

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		total.Add(1)

		if r.Header.Get("Via") == "" {
			direct.Add(1)
		}

		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `<html><a href="/a">a</a> <a href="/redirect">redirect</a></html>`)
		case "/a":
			fmt.Fprint(w, `<html><a href="/b">b</a></html>`)
		case "/redirect":
			http.Redirect(w, r, "/landed", http.StatusFound)
		case "/sitemap.xml":
			w.Header().Set("Content-Type", "application/xml")
			fmt.Fprint(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/from-sitemap</loc></url></urlset>`)
		default:
			fmt.Fprint(w, "<html></html>")
		}
	})

	counts := make([]atomic.Int32, 3)
	proxies := make([]string, len(counts))

	for i := range proxies {
		proxies[i] = newTestProxy(t, &counts[i]).URL
	}

	crawlWith := func(proxies []string) []string {
		total.Store(0)
		direct.Store(0)

		for i := range counts {
			counts[i].Store(0)
		}

		results := crawl(t, &crawler.Scan{
			Target:  server.URL,
			Proxies: proxies,
			Threads: testThreads,
			Timeout: input.TimeoutRequest,
		})

		return slices.Sorted(slices.Values(results.URLs))
	}

	want := crawlWith(nil)
	wantRequests := total.Load()

	if !slices.Contains(want, server.URL+"/b") || !slices.Contains(want, server.URL+"/from-sitemap") {
		t.Fatalf("without proxies got %v", want)
	}

	for _, n := range []int{1, 3} {
		got := crawlWith(proxies[:n])

		if !slices.Equal(got, want) {
			t.Errorf("with %d proxies got %v, want %v", n, got, want)
		}

		if direct.Load() != 0 || total.Load() != wantRequests {
			t.Errorf("with %d proxies %d of %d requests were direct, want 0 of %d",
				n, direct.Load(), total.Load(), wantRequests)
		}

		// Each proxy has its turn: the requests forwarded differ at most by one.
		forwarded := make([]int32, len(counts))
		for i := range counts {
			forwarded[i] = counts[i].Load()
		}

		used, unused := forwarded[:n], forwarded[n:]
		if slices.Max(used)-slices.Min(used) > 1 || slices.ContainsFunc(unused, func(c int32) bool { return c != 0 }) {
			t.Errorf("with %d proxies the requests forwarded are %v", n, forwarded)
		}
	}
}

// TestNewProxiesTargets checks that the turn of the proxies goes on from a
// target to the next one: starting again from the first proxy, the first
// requests of every target would always use the same proxies.
func TestNewProxiesTargets(t *testing.T) {
	counts := make([]atomic.Int32, 3)
	proxies := make([]string, len(counts))

	for i := range proxies {
		proxies[i] = newTestProxy(t, &counts[i]).URL
	}

	// Each target is requested twice: the sitemap probe and the target.
	for range len(proxies) {
		server := serve(t, func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "<html></html>")
		})

		crawl(t, &crawler.Scan{
			Target:  server.URL,
			Proxies: proxies,
			Threads: testThreads,
			Timeout: input.TimeoutRequest,
		})
	}

	for i := range counts {
		if got := counts[i].Load(); got != 2 {
			t.Errorf("proxy %d forwarded %d requests, want 2", i, got)
		}
	}
}

// TestNewProxiesReuse checks that the connections to the proxies are reused
// with more proxies than parallel requests too: each proxy opens a single
// connection, as a single proxy does.
func TestNewProxiesReuse(t *testing.T) {
	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			for i := range 10 {
				fmt.Fprintf(w, `<a href="/%d">%d</a> `, i, i)
			}
		}
	})

	for _, n := range []int{1, 2, 3} {
		requests := make([]atomic.Int32, n)
		conns := make([]atomic.Int32, n)
		proxies := make([]string, n)

		for i := range proxies {
			proxy := httptest.NewUnstartedServer(proxyHandler(t, &requests[i]))
			proxy.Config.ConnState = func(_ net.Conn, state http.ConnState) {
				if state == http.StateNew {
					conns[i].Add(1)
				}
			}

			proxy.Start()
			t.Cleanup(proxy.Close)

			proxies[i] = proxy.URL
		}

		// One request at a time: the requests use the proxies in turn.
		crawl(t, &crawler.Scan{
			Target:  server.URL,
			Proxies: proxies,
			Threads: 1,
			Timeout: input.TimeoutRequest,
		})

		for i := range proxies {
			// More requests than one: the connection was reused.
			if requests[i].Load() < 2 || conns[i].Load() != 1 {
				t.Errorf("with %d proxies, proxy %d forwarded %d requests with %d connections, want 1 connection",
					n, i, requests[i].Load(), conns[i].Load())
			}
		}
	}
}

// TestNewProxiesDebug checks that in debug mode the requests failed show
// the proxy used, without its password, and that the other proxies are used.
func TestNewProxiesDebug(t *testing.T) {
	server := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `<html><a href="/a">a</a> <a href="/b">b</a></html>`)
	})

	var forwarded atomic.Int32

	working := newTestProxy(t, &forwarded)

	down := httptest.NewServer(http.NotFoundHandler())
	down.Close()

	var logs bytes.Buffer

	previous := log.Writer()
	defer log.SetOutput(previous)

	log.SetOutput(&logs)

	crawl(t, &crawler.Scan{
		Target:  server.URL,
		Proxies: []string{working.URL, strings.Replace(down.URL, "http://", "http://user:secret@", 1)},
		Debug:   true,
		Threads: testThreads,
		Timeout: input.TimeoutRequest,
	})

	want := "proxy " + strings.Replace(down.URL, "http://", "http://user:xxxxx@", 1) + ": "
	if !strings.Contains(logs.String(), want) || strings.Contains(logs.String(), "secret") {
		t.Errorf("debug output %q, want the requests failed with %q", logs.String(), want)
	}

	if forwarded.Load() == 0 {
		t.Error("the working proxy was not used")
	}
}

// TestNewTransportProxiesOnly checks that with more proxies a request is
// never sent directly to the target: only through the proxy chosen.
func TestNewTransportProxiesOnly(t *testing.T) {
	var direct atomic.Int32

	server := serve(t, func(_ http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Via") == "" {
			direct.Add(1)
		}
	})

	var forwarded atomic.Int32

	proxy := newTestProxy(t, &forwarded).URL

	transport, roundTripper := crawler.NewTransport(1, []string{proxy, proxy})
	defer transport.CloseIdleConnections()

	for _, rt := range []http.RoundTripper{roundTripper, transport} {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, server.URL, nil)
		if err != nil {
			t.Fatal(err)
		}

		resp, err := rt.RoundTrip(req)
		if err == nil {
			resp.Body.Close()
		}

		// Only through the round tripper, which chooses the proxy.
		if (err == nil) != (rt == roundTripper) {
			t.Errorf("request through %T: error %v", rt, err)
		}
	}

	if direct.Load() != 0 || forwarded.Load() != 1 {
		t.Errorf("%d direct requests and %d through the proxy, want 0 and 1", direct.Load(), forwarded.Load())
	}
}
