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
	"slices"
	"strings"
	"testing"

	"github.com/edoardottt/cariddi/pkg/crawler"
	"github.com/edoardottt/cariddi/pkg/input"
)

const nextPage = "/next"

// TestMetaRefreshURL checks the parsing of meta refresh redirects, as
// browsers do it.
func TestMetaRefreshURL(t *testing.T) {
	tests := []struct {
		content string
		want    string // "" if there's no URL.
	}{
		{"0; url=/next", nextPage},
		{"0;URL=/next", nextPage},
		{"0; URL = /next", nextPage},
		{"0, url=/next", nextPage},
		{"0 url=/next", nextPage},
		{"\t5 ;\n url=/next ", nextPage},
		{"0;url='/next'", nextPage},
		{`0;url="/next"`, nextPage},
		{"0;url='/next' and more", nextPage},
		{"0;url='/next", nextPage},
		{"0; /next", nextPage},
		{"0,/next", nextPage},
		{"0; '/next'", nextPage},
		{".5;url=/next", nextPage},
		{"1.5; url=/next", nextPage},
		{"url=/next", nextPage}, // no delay: ignored by browsers, accepted.
		{"0; url=/a?b=1;c=2", "/a?b=1;c=2"},
		{"0; url=https://other.com/x", "https://other.com/x"},
		{"0; update.html", "update.html"}, // not url=, but starts with u.
		{"0; url.html", "url.html"},
		{"5", ""},
		{"", ""},
		{"0;", ""},
		{"0; url=", ""},
		{"0; url=''", ""},
		{"5x; url=/next", ""},
		{"/next", ""},
		{"text/html; charset=utf-8", ""},
	}

	for _, tt := range tests {
		got, ok := crawler.MetaRefreshURL("refresh", tt.content)
		if got != tt.want || ok != (tt.want != "") {
			t.Errorf("MetaRefreshURL(refresh, %q) = %q, %v; want %q", tt.content, got, ok, tt.want)
		}
	}

	for _, httpEquiv := range []string{"Refresh", "REFRESH", " refresh "} {
		if got, ok := crawler.MetaRefreshURL(httpEquiv, "0; url=/next"); !ok || got != nextPage {
			t.Errorf("MetaRefreshURL(%q) = %q, %v; want %q", httpEquiv, got, ok, nextPage)
		}
	}

	if _, ok := crawler.MetaRefreshURL("content-type", "0; url=/next"); ok {
		t.Error("MetaRefreshURL accepted http-equiv content-type")
	}
}

// TestNewMetaRefresh checks that the meta refresh redirects are followed.
func TestNewMetaRefresh(t *testing.T) {
	pages := map[string]string{
		"head":          `<head><meta http-equiv="refresh" content="0; url=/t/head"></head>`,
		"case":          `<head><META HTTP-EQUIV="Refresh" CONTENT="0;URL=/t/case"></head>`,
		"body":          `<body><meta http-equiv="refresh" content="3,/t/body"></body>`,
		"quotes":        `<meta http-equiv="refresh" content="0; URL = '/t/quotes'">`,
		"semicolon":     `<meta http-equiv="refresh" content="0; url=/t/semicolon?a=1;b=2">`,
		"no-delay":      `<meta http-equiv="refresh" content="url=/t/no-delay">`,
		"noscript-head": `<head><noscript><meta http-equiv="refresh" content="0; url=/t/noscript-head"></noscript></head>`,
		"noscript-body": `<body><noscript><meta http-equiv="refresh" content="0;url=/t/noscript-body"></noscript></body>`,
		"second-meta": `<meta http-equiv="refresh" content="0; url=/t/first">` +
			`<meta http-equiv="refresh" content="9; url=/t/second-meta">`,
	}
	targets := []string{"/t/first", "/t/semicolon?a=1;b=2"}

	for name := range pages {
		if name != "semicolon" {
			targets = append(targets, "/t/"+name)
		}
	}

	server := serve(t, func(w http.ResponseWriter, r *http.Request) {
		page := pages[strings.TrimPrefix(r.URL.Path, "/p/")]

		if r.URL.Path == "/" {
			for name := range pages {
				page += fmt.Sprintf(`<a href="/p/%s">%s</a>`, name, name)
			}
		}

		fmt.Fprintf(w, "<html>%s</html>", page)
	})

	results := crawl(t, &crawler.Scan{
		Target:      server.URL,
		Concurrency: testConcurrency,
		Timeout:     input.TimeoutRequest,
	})

	for _, target := range targets {
		if !slices.Contains(results.URLs, server.URL+target) {
			t.Errorf("meta refresh to %s not followed", target)
		}
	}
}
