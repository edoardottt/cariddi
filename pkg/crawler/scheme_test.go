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
		name      string
		mode      string
		reachable bool
		expected  string
	}{
		{"http mode forces http", input.SchemeHTTP, true, "http"},
		{"https mode forces https", input.SchemeHTTPS, false, "https"},
		{"auto uses https when reachable", input.SchemeAuto, true, "https"},
		{"auto falls back to http when unreachable", input.SchemeAuto, false, "http"},
		{"empty mode behaves like auto (reachable)", "", true, "https"},
		{"empty mode behaves like auto (unreachable)", "", false, "http"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := crawler.ResolveProtocol(tt.mode, func() bool { return tt.reachable })
			if got != tt.expected {
				t.Errorf("ResolveProtocol(mode=%q, reachable=%v) = %q; want %q",
					tt.mode, tt.reachable, got, tt.expected)
			}
		})
	}
}
