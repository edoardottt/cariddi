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

package crawler

import (
	"bytes"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	// maxRedirects is the maximum number of redirects followed (as Go does).
	maxRedirects = 10
	// sitemapProbeKey marks, in the request context, the sitemap.xml
	// request sent to the target.
	sitemapProbeKey = "sitemapProbe"
	// htmlSpace is the ASCII whitespace of the HTML standard.
	htmlSpace = "\t\n\f\r "
	// idleConnTimeout is how long an unused connection is kept open
	// (as in http.DefaultTransport).
	idleConnTimeout = 90 * time.Second
)

// RetrieveBody retrieves the body (in the response) of a url.
func RetrieveBody(target *string) string {
	sb, err := GetRequest(*target)
	if err == nil && sb != "" {
		return sb
	}

	return ""
}

// IgnoreMatch checks if the URL should be ignored or not.
func IgnoreMatch(url string, ignoreSlice *[]string) bool {
	for _, ignore := range *ignoreSlice {
		if strings.Contains(url, ignore) {
			return true
		}
	}

	return false
}

// canVisit checks if a URL can be crawled: it must be
// in scope and not ignored.
func canVisit(event *Event, absoluteURL string) bool {
	return event.Scope.Contains(absoluteURL) &&
		(!event.Ignore || !IgnoreMatch(absoluteURL, &event.IgnoreSlice))
}

// nextLinks collects the links to crawl at the next depth, each one once.
type nextLinks struct {
	mu     sync.Mutex
	links  []string
	queued map[string]struct{}
}

// add queues link, unless it was already queued (at this or a lower depth).
func (n *nextLinks) add(link string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if _, ok := n.queued[link]; ok {
		return
	}

	if n.queued == nil {
		n.queued = map[string]struct{}{}
	}

	n.queued[link] = struct{}{}
	n.links = append(n.links, link)
}

// take returns the queued links and empties the queue.
func (n *nextLinks) take() []string {
	n.mu.Lock()
	defer n.mu.Unlock()

	links := n.links
	n.links = nil

	return links
}

// isSuccess checks if r is a successful (2xx) response.
func isSuccess(r *colly.Response) bool {
	return r.StatusCode >= http.StatusOK && r.StatusCode < http.StatusMultipleChoices
}

// isSitemapProbe checks if r is the sitemap.xml request sent to the target
// (the requests for the links found in it share its context, but are deeper,
// or, with a max depth, are sent with a new context).
func isSitemapProbe(r *colly.Request) bool {
	return r.Depth == 1 && r.Ctx.Get(sitemapProbeKey) != ""
}

// isSitemap checks if body is an XML sitemap (or sitemap index).
func isSitemap(body []byte) bool {
	return bytes.Contains(body, []byte("<urlset")) || bytes.Contains(body, []byte("<sitemapindex"))
}

// metaRefreshURL returns the URL of a meta refresh redirect, given its
// http-equiv (in any case) and content attributes. The content is parsed
// as browsers do (HTML standard, "shared declarative refresh steps"), e.g.
// `5; URL='/next'` and `0,/next` -> /next, `0; url=/a?b=1;c=2` -> /a?b=1;c=2,
// and `url=/next` is accepted without the delay too.
func metaRefreshURL(httpEquiv, content string) (string, bool) {
	if !strings.EqualFold(strings.Trim(httpEquiv, htmlSpace), "refresh") {
		return "", false
	}

	s := strings.TrimLeft(content, htmlSpace)

	// The delay (e.g. 0, 5 or 1.5) is ignored.
	rest := strings.TrimLeft(s, "0123456789.")

	switch {
	case rest == s: // no delay: accept only an explicit "url=".
		if _, ok := trimURLPrefix(rest); !ok {
			return "", false
		}
	case rest != "":
		// The delay must be followed by ";", "," or a space.
		if !strings.ContainsRune(";,"+htmlSpace, rune(rest[0])) {
			return "", false
		}

		rest = strings.TrimLeft(rest, htmlSpace)
		if rest != "" && (rest[0] == ';' || rest[0] == ',') {
			rest = rest[1:]
		}

		rest = strings.TrimLeft(rest, htmlSpace)
	}

	if trimmed, ok := trimURLPrefix(rest); ok {
		rest = trimmed
	} else if rest != "" && (rest[0] == 'u' || rest[0] == 'U') {
		// Not "url=", but a URL starting with u (e.g. update.html).
		rest = strings.TrimRight(rest, htmlSpace)

		return rest, rest != ""
	}

	// Quoted URL, e.g. url='/next': the closing quote is optional.
	if rest != "" && (rest[0] == '\'' || rest[0] == '"') {
		quote := rest[0]

		rest = rest[1:]
		if i := strings.IndexByte(rest, quote); i >= 0 {
			rest = rest[:i]
		}
	}

	rest = strings.Trim(rest, htmlSpace)

	return rest, rest != ""
}

// trimURLPrefix removes "url=" (in any case, with spaces around "=")
// from the start of s.
func trimURLPrefix(s string) (string, bool) {
	if len(s) < len("url") || !strings.EqualFold(s[:len("url")], "url") {
		return s, false
	}

	rest := strings.TrimLeft(s[len("url"):], htmlSpace)
	if !strings.HasPrefix(rest, "=") {
		return s, false
	}

	return strings.TrimLeft(rest[1:], htmlSpace), true
}
