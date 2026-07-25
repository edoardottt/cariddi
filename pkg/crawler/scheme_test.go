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
	"testing"

	"github.com/edoardottt/cariddi/pkg/crawler"
	"github.com/edoardottt/cariddi/pkg/input"
)

func TestResolveProtocol(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		mode       string
		httpsUp    bool
		httpUp     bool
		wantScheme string
		wantOK     bool
	}{
		{"auto prefers https when both up", input.SchemeAuto, true, true, "https", true},
		{"auto uses https when only https up", input.SchemeAuto, true, false, "https", true},
		{"auto falls back to http when only http up", input.SchemeAuto, false, true, "http", true},
		{"auto skips when nothing up", input.SchemeAuto, false, false, "", false},
		{"empty mode behaves like auto", "", false, true, "http", true},
		{"https crawls when 443 up", input.SchemeHTTPS, true, false, "https", true},
		{"https skips when 443 down", input.SchemeHTTPS, false, true, "", false},
		{"http crawls when 80 up", input.SchemeHTTP, false, true, "http", true},
		{"http skips when 80 down", input.SchemeHTTP, true, false, "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			reachable := func(scheme string) bool {
				switch scheme {
				case input.SchemeHTTPS:
					return tt.httpsUp
				case input.SchemeHTTP:
					return tt.httpUp
				default:
					return false
				}
			}

			gotScheme, gotOK := crawler.ResolveProtocol(tt.mode, reachable)
			if gotScheme != tt.wantScheme || gotOK != tt.wantOK {
				t.Errorf("ResolveProtocol(mode=%q, httpsUp=%v, httpUp=%v) = (%q, %v); want (%q, %v)",
					tt.mode, tt.httpsUp, tt.httpUp, gotScheme, gotOK, tt.wantScheme, tt.wantOK)
			}
		})
	}
}
