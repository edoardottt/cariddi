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
	"strings"

	"github.com/gocolly/colly/v2"
)

const (
	// maxRedirects is the maximum number of redirects followed (as Go does).
	maxRedirects = 10
	// sitemapProbeKey marks, in the request context, the sitemap.xml
	// request sent to the target.
	sitemapProbeKey = "sitemapProbe"
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

// isSitemapProbe checks if r is the sitemap.xml request sent to the target
// (the requests for the links found in it share its context, but are deeper).
func isSitemapProbe(r *colly.Request) bool {
	return r.Depth == 1 && r.Ctx.Get(sitemapProbeKey) != ""
}

// isSitemap checks if body is an XML sitemap (or sitemap index).
func isSitemap(body []byte) bool {
	return bytes.Contains(body, []byte("<urlset")) || bytes.Contains(body, []byte("<sitemapindex"))
}
