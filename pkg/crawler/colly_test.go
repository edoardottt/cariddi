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
	"os"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/edoardottt/cariddi/pkg/crawler"
	"github.com/edoardottt/cariddi/pkg/input"
)

const (
	testPages       = 200
	testConcurrency = 20
)

// awsKey returns a distinct, well-formed AWS access key for page i.
func awsKey(i int) string {
	return fmt.Sprintf("AKIA%016d", i)
}

// newTestSite serves an index linking to testPages pages, each one with a
// distinct secret and an error message.
func newTestSite(t *testing.T) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			var sb strings.Builder

			sb.WriteString("<html><body>\n")

			for i := range testPages {
				fmt.Fprintf(&sb, "<a href=\"/page/%d\">page %d</a>\n", i, i)
			}

			sb.WriteString("</body></html>\n")
			fmt.Fprint(w, sb.String())

			return
		}

		var i int
		if _, err := fmt.Sscanf(r.URL.Path, "/page/%d", &i); err != nil {
			http.NotFound(w, r)

			return
		}

		fmt.Fprintf(w, "<html><body><script>var k = \"%s\";</script><p>fatal error</p></body></html>\n", awsKey(i))
	}))
	t.Cleanup(server.Close)

	return server
}

// crawl runs the crawler on the test site, with stdout silenced.
func crawl(t *testing.T, scan *crawler.Scan) *crawler.Results {
	t.Helper()

	return crawlContext(t, t.Context(), scan)
}

// crawlContext is crawl, stopped when ctx is canceled.
func crawlContext(t *testing.T, ctx context.Context, scan *crawler.Scan) *crawler.Results {
	t.Helper()

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}

	stdout := os.Stdout
	os.Stdout = devNull

	defer func() {
		os.Stdout = stdout

		devNull.Close()
	}()

	return crawler.New(ctx, scan)
}

// TestNewFindsEverySecretAndError checks that the results collected by the
// concurrent colly callbacks are complete (run it with -race too).
func TestNewFindsEverySecretAndError(t *testing.T) {
	server := newTestSite(t)
	results := crawl(t, &crawler.Scan{
		Target:      server.URL,
		SecretsFlag: true,
		ErrorsFlag:  true,
		Concurrency: testConcurrency,
		Timeout:     input.TimeoutRequest,
	})

	secrets := map[string]string{}
	for _, s := range results.Secrets {
		if _, ok := secrets[s.Match]; ok {
			t.Errorf("secret %q found twice", s.Match)
		}

		secrets[s.Match] = s.URL
	}

	errorPages := map[string]bool{}
	for _, e := range results.Errors {
		if errorPages[e.URL] {
			t.Errorf("error found twice in %s", e.URL)
		}

		errorPages[e.URL] = true
	}

	for i := range testPages {
		url := fmt.Sprintf("%s/page/%d", server.URL, i)

		if got, ok := secrets[awsKey(i)]; !ok || got != url {
			t.Errorf("secret of page %d: got URL %q (found: %v)", i, got, ok)
		}

		if !errorPages[url] {
			t.Errorf("error of page %d not found", i)
		}
	}

	if len(results.Secrets) != testPages || len(results.Errors) != testPages {
		t.Errorf("got %d secrets and %d errors, want %d of each", len(results.Secrets), len(results.Errors), testPages)
	}
}

// TestNewCustomSecrets checks the regexes of a secrets file: blank lines
// and invalid regexes are skipped.
func TestNewCustomSecrets(t *testing.T) {
	server := newTestSite(t)
	results := crawl(t, &crawler.Scan{
		Target:       server.URL,
		SecretsFlag:  true,
		SecretsSlice: []string{"AKIA0{14}42", "", "   ", "[invalid"},
		Concurrency:  testConcurrency,
		Timeout:      input.TimeoutRequest,
	})

	if len(results.Secrets) != 1 || results.Secrets[0].Match != awsKey(42) ||
		results.Secrets[0].Secret.Name != "CustomFromFile" {
		t.Fatalf("got %v, want only %s", results.Secrets, awsKey(42))
	}
}

// TestNewBlankSecretsFile checks that a secrets file with only blank lines
// is like an empty one: the predefined regexes are used.
func TestNewBlankSecretsFile(t *testing.T) {
	server := newTestSite(t)
	results := crawl(t, &crawler.Scan{
		Target:       server.URL,
		SecretsFlag:  true,
		SecretsSlice: []string{"", "  ", "\t"},
		Concurrency:  testConcurrency,
		Timeout:      input.TimeoutRequest,
	})

	if len(results.Secrets) != testPages {
		t.Fatalf("got %d secrets, want %d", len(results.Secrets), testPages)
	}

	for _, s := range results.Secrets {
		if s.Secret.Name != "AWS Access Key" {
			t.Fatalf("unexpected secret %q: %q", s.Secret.Name, s.Match)
		}
	}
}

// serve starts a test server closed at the end of the test.
func serve(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server
}

// TestNewRedirects checks that only the redirects in scope are followed.
func TestNewRedirects(t *testing.T) {
	var outside, landed atomic.Int32

	other := serve(t, func(_ http.ResponseWriter, _ *http.Request) {
		outside.Add(1)
	})

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `<html><a href="/in">in</a> <a href="/out">out</a></html>`)
		case "/in":
			http.Redirect(w, r, "/landed", http.StatusFound)
		case "/out":
			http.Redirect(w, r, other.URL+"/login", http.StatusFound)
		case "/landed":
			landed.Add(1)
		default:
			http.NotFound(w, r)
		}
	})

	crawl(t, &crawler.Scan{
		Target:      server.URL,
		Concurrency: testConcurrency,
		Timeout:     input.TimeoutRequest,
	})

	if got := landed.Load(); got != 1 {
		t.Errorf("redirect in scope followed %d times, want 1", got)
	}

	if got := outside.Load(); got != 0 {
		t.Errorf("host out of scope got %d requests, want 0", got)
	}
}

// TestNewMaxDepth checks that the links are followed up to the max depth,
// the target is at depth 0, its links at depth 1.
func TestNewMaxDepth(t *testing.T) {
	const chainLength = 5

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		var i int
		if _, err := fmt.Sscanf(r.URL.Path, "/chain/%d", &i); (err != nil && r.URL.Path != "/") || i >= chainLength {
			http.NotFound(w, r)

			return
		}

		fmt.Fprintf(w, `<html><a href="/chain/%d">next</a></html>`, i+1)
	})

	for maxDepth := 1; maxDepth <= 3; maxDepth++ {
		results := crawl(t, &crawler.Scan{
			Target:      server.URL,
			MaxDepth:    maxDepth,
			Concurrency: testConcurrency,
			Timeout:     input.TimeoutRequest,
		})

		last := fmt.Sprintf("%s/chain/%d", server.URL, maxDepth)
		next := fmt.Sprintf("%s/chain/%d", server.URL, maxDepth+1)

		if !slices.Contains(results.URLs, last) || slices.Contains(results.URLs, next) {
			t.Errorf("max depth %d: got %v, want %s but not %s", maxDepth, results.URLs, last, next)
		}
	}
}

// TestNewMaxDepthLowestDepth checks that a page gets its lowest depth (the
// fewest links from the target, whatever its URL path), even if a longer
// chain of links to it is found first: /x/y/hub is linked by the target
// (depth 1) after many other links, and by /b (which would make it depth 2).
func TestNewMaxDepthLowestDepth(t *testing.T) {
	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `<html><a href="/b">b</a>`)

			for i := range testPages * 5 {
				fmt.Fprintf(w, `<a href="/page/%d">page</a>`, i)
			}

			fmt.Fprint(w, `<a href="/x/y/hub">hub</a></html>`)
		case "/b":
			fmt.Fprint(w, `<html><a href="/x/y/hub">hub</a></html>`)
		case "/x/y/hub":
			fmt.Fprint(w, `<html><a href="/leaf">leaf</a></html>`)
		default:
			fmt.Fprint(w, "<html></html>")
		}
	})

	results := crawl(t, &crawler.Scan{
		Target:      server.URL,
		MaxDepth:    2,
		Concurrency: testConcurrency,
		Timeout:     input.TimeoutRequest,
	})

	// The target (depth 0) -> /x/y/hub (depth 1) -> /leaf (depth 2).
	if !slices.Contains(results.URLs, server.URL+"/leaf") {
		t.Errorf("max depth 2: %s/leaf (depth 2) not crawled", server.URL)
	}
}

// TestNewSitemapProbe checks that the sitemap.xml requested by default is
// in the results only if it's a valid sitemap.
func TestNewSitemapProbe(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		sitemap string
		want    bool
	}{
		{name: "missing", status: http.StatusNotFound, sitemap: "not found", want: false},
		{name: "not a sitemap", status: http.StatusOK, sitemap: "<html>soft 404</html>", want: false},
		{
			name:   "valid",
			status: http.StatusOK,
			sitemap: `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9"><url><loc>/from-sitemap</loc></url></urlset>`,
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := serve(t, func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/sitemap.xml":
					w.Header().Set("Content-Type", "application/xml")
					w.WriteHeader(tt.status)
					fmt.Fprint(w, tt.sitemap)
				default:
					fmt.Fprint(w, "<html></html>")
				}
			})

			results := crawl(t, &crawler.Scan{
				Target:      server.URL,
				Concurrency: testConcurrency,
				Timeout:     input.TimeoutRequest,
			})

			if got := slices.Contains(results.URLs, server.URL+"/sitemap.xml"); got != tt.want {
				t.Errorf("sitemap in results = %v, want %v (results: %v)", got, tt.want, results.URLs)
			}

			if got := slices.Contains(results.URLs, server.URL+"/from-sitemap"); got != tt.want {
				t.Errorf("link of the sitemap in results = %v, want %v (results: %v)", got, tt.want, results.URLs)
			}
		})
	}
}

// TestNewErrorPages checks that the error pages are scanned, but their links
// are not followed and their URLs are not findings (endpoints, extensions).
func TestNewErrorPages(t *testing.T) {
	var linked atomic.Int32

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			fmt.Fprint(w, `<html><a href="/backup.sql?id=1">a</a> <a href="/crash">b</a></html>`)
		case "/backup.sql":
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `<html>%s <a href="/linked">x</a></html>`, awsKey(404))
		case "/crash":
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "<html>Fatal error: Uncaught exception</html>")
		case "/linked":
			linked.Add(1)
		default:
			http.NotFound(w, r)
		}
	})

	results := crawl(t, &crawler.Scan{
		Target:        server.URL,
		SecretsFlag:   true,
		ErrorsFlag:    true,
		EndpointsFlag: true,
		FileType:      7,
		Concurrency:   testConcurrency,
		Timeout:       input.TimeoutRequest,
	})

	if len(results.Secrets) != 1 || results.Secrets[0].Match != awsKey(404) {
		t.Errorf("got secrets %v, want %s (404 page)", results.Secrets, awsKey(404))
	}

	if len(results.Errors) == 0 || results.Errors[0].URL != server.URL+"/crash" {
		t.Errorf("got errors %v, want the errors of the 500 page", results.Errors)
	}

	if len(results.Endpoints) != 0 || len(results.Extensions) != 0 {
		t.Errorf("got endpoints %v and extensions %v, want none", results.Endpoints, results.Extensions)
	}

	if got := linked.Load(); got != 0 {
		t.Errorf("link of the 404 page requested %d times, want 0", got)
	}
}

// TestNewCanceled checks that a canceled crawl (CTRL+C) returns at once, even
// with a delay between the requests, with the results found so far, and that
// the request in flight is aborted.
func TestNewCanceled(t *testing.T) {
	const pages = 20

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	var requested atomic.Int32

	aborted := make(chan struct{})

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/index" {
			fmt.Fprintf(w, "<html>%s", awsKey(0))

			for i := 1; i <= pages; i++ {
				fmt.Fprintf(w, `<a href="/page/%d">page</a>`, i)
			}

			fmt.Fprint(w, "</html>")

			return
		}

		// The first page requested cancels the crawl and never answers.
		if requested.Add(1) == 1 {
			cancel()
			<-r.Context().Done()
			close(aborted)
		}
	})

	start := time.Now()
	results := crawlContext(t, ctx, &crawler.Scan{
		Target:      server.URL + "/index", // not the root: no sitemap probe.
		SecretsFlag: true,
		// Nothing is printed by the requests left, which can still run
		// after New returns (when stdout is restored).
		JSON:        true,
		Delay:       1,
		Concurrency: 1,
		Timeout:     input.TimeoutRequest,
	})

	// Waiting for the requests left would take a delay each (20s).
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Errorf("canceled crawl returned after %v", elapsed)
	}

	if len(results.Secrets) != 1 || results.Secrets[0].Match != awsKey(0) {
		t.Errorf("got secrets %v, want %s (found before the cancel)", results.Secrets, awsKey(0))
	}

	select {
	case <-aborted:
	case <-time.After(5 * time.Second):
		t.Error("the request in flight was not aborted")
	}
}

// TestNewTargets checks the parsing of the targets: http is used if there's
// no scheme (even if "://" is in the query), the case of the path is kept and
// the invalid targets are skipped.
func TestNewTargets(t *testing.T) {
	server := serve(t, func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "<html></html>")
	})
	host := strings.TrimPrefix(server.URL, "http://")

	tests := []struct {
		target    string
		intensive bool
		want      []string
	}{
		{target: host + "/?next=https://example.org", want: []string{server.URL + "/?next=https://example.org"}},
		{target: "HTTP://" + host + "/Index.HTML", want: []string{server.URL + "/Index.HTML"}},
		{target: "http://"},
		{target: "https://co.uk/", intensive: true}, // a TLD.
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			results := crawl(t, &crawler.Scan{
				Target:      tt.target,
				Intensive:   tt.intensive,
				Concurrency: testConcurrency,
				Timeout:     input.TimeoutRequest,
			})

			if !slices.Equal(results.URLs, tt.want) {
				t.Errorf("got %v, want %v", results.URLs, tt.want)
			}
		})
	}
}

// TestNewRequestsEachURLOnce checks that a URL found at the same time in
// different pages is requested only once.
func TestNewRequestsEachURLOnce(t *testing.T) {
	const pages = 1000

	var requests sync.Map

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		count, _ := requests.LoadOrStore(r.URL.Path, new(atomic.Int32))
		count.(*atomic.Int32).Add(1)

		var i int
		if _, err := fmt.Sscanf(r.URL.Path, "/page/%d", &i); err != nil && r.URL.Path != "/" {
			http.NotFound(w, r)

			return
		}

		var sb strings.Builder
		for k := 1; k <= 10; k++ {
			fmt.Fprintf(&sb, `<a href="/page/%d">x</a>`, (i*7+k)%pages)
		}

		fmt.Fprintf(w, "<html>%s</html>", sb.String())
	})

	crawl(t, &crawler.Scan{
		Target:      server.URL,
		Concurrency: testConcurrency,
		Timeout:     input.TimeoutRequest,
	})

	requests.Range(func(path, count any) bool {
		if n := count.(*atomic.Int32).Load(); n != 1 {
			t.Errorf("%s requested %d times, want 1", path, n)
		}

		return true
	})
}

// TestNewSitemapIndex checks that the sitemaps listed in a sitemap index
// are crawled, in scope only and once, up to the max depth.
func TestNewSitemapIndex(t *testing.T) {
	tests := []struct {
		name     string
		maxDepth int
		wantPage bool
	}{
		{name: "no max depth", maxDepth: 0, wantPage: true},
		// index (depth 0) -> sitemap (depth 1) -> page (depth 2)
		{name: "max depth 1", maxDepth: 1, wantPage: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var outside atomic.Int32

			other := serve(t, func(_ http.ResponseWriter, _ *http.Request) {
				outside.Add(1)
			})

			var requests sync.Map

			server := serve(t, func(w http.ResponseWriter, r *http.Request) {
				count, _ := requests.LoadOrStore(r.URL.Path, new(atomic.Int32))
				count.(*atomic.Int32).Add(1)

				w.Header().Set("Content-Type", "application/xml")

				switch r.URL.Path {
				case "/sitemap.xml":
					// It lists itself too, to check that loops end.
					fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <sitemap>
    <loc>
      http://%[1]s/sitemap-pages.xml
    </loc>
    <lastmod>2026-10-26</lastmod>
  </sitemap>
  <sitemap><loc>http://%[1]s/sitemap.xml</loc></sitemap>
  <sitemap><loc>%[2]s/sitemap-outside.xml</loc></sitemap>
</sitemapindex>`, r.Host, other.URL)
				case "/sitemap-pages.xml":
					fmt.Fprintf(w, `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url><loc>http://%s/from-sitemap</loc><lastmod>2026-10-26</lastmod></url>
</urlset>`, r.Host)
				default:
					w.Header().Set("Content-Type", "text/html")
					fmt.Fprint(w, "<html></html>")
				}
			})

			results := crawl(t, &crawler.Scan{
				Target:      server.URL,
				MaxDepth:    tt.maxDepth,
				Concurrency: testConcurrency,
				Timeout:     input.TimeoutRequest,
			})

			for _, path := range []string{"/sitemap.xml", "/sitemap-pages.xml"} {
				if !slices.Contains(results.URLs, server.URL+path) {
					t.Errorf("%s not in results: %v", path, results.URLs)
				}
			}

			if got := slices.Contains(results.URLs, server.URL+"/from-sitemap"); got != tt.wantPage {
				t.Errorf("page of the listed sitemap in results = %v, want %v", got, tt.wantPage)
			}

			if got := outside.Load(); got != 0 {
				t.Errorf("host out of scope got %d requests, want 0", got)
			}

			requests.Range(func(path, count any) bool {
				if n := count.(*atomic.Int32).Load(); n != 1 {
					t.Errorf("%s requested %d times, want 1", path, n)
				}

				return true
			})
		})
	}
}
