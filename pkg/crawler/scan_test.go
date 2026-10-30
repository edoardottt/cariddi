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
	"net/url"
	"testing"

	"github.com/edoardottt/cariddi/pkg/crawler"
)

// TestHuntExtensions checks that the extension is hunted in the decoded
// path only (not in the query or the fragment).
func TestHuntExtensions(t *testing.T) {
	tests := []struct {
		input    string
		severity int
		want     string // "" if no extension is found.
	}{
		{"https://example.com/backup.sql", 7, "sql"},
		{"https://example.com/BACKUP.SQL", 7, "sql"},
		{"https://example.com/backup.sql#top", 7, "sql"},
		{"https://example.com/backup%2Esql", 7, "sql"},
		{"https://example.com/a.tar.gz", 7, "tar.gz"},
		{"https://example.com/page?file=backup.sql", 7, ""},
		{"https://example.com/backup.sql/", 7, ""},
		{"https://example.com/index.html", 6, "html"},
		{"https://example.com/index.html", 5, ""},
	}

	for _, tt := range tests {
		u, err := url.Parse(tt.input)
		if err != nil {
			t.Fatal(err)
		}

		got := crawler.HuntExtensions(u, tt.severity)
		if got.Filetype.Extension != tt.want {
			t.Errorf("HuntExtensions(%s, %d) = %q, want %q", tt.input, tt.severity, got.Filetype.Extension, tt.want)
		}

		// The URL found is the one crawled, as it is.
		if tt.want != "" && got.URL != tt.input {
			t.Errorf("HuntExtensions(%s, %d) URL = %s", tt.input, tt.severity, got.URL)
		}
	}
}
