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
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

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

	return crawler.New(scan)
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
