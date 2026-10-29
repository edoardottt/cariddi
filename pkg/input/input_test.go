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

package input_test

import (
	"maps"
	"testing"

	"github.com/edoardottt/cariddi/pkg/input"
)

func TestGetHeaders(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]string
	}{
		{
			name:  "flag format",
			input: "Cookie: auth=yes;;Client: type=2",
			want:  map[string]string{"Cookie": "auth=yes", "Client": "type=2"},
		},
		{
			name:  "one per line",
			input: "Cookie: auth=yes\nClient: type=2\n",
			want:  map[string]string{"Cookie": "auth=yes", "Client": "type=2"},
		},
		{
			name:  "crlf, blank lines and both separators",
			input: "\r\nCookie: a=1; b=2\r\n\r\nReferer: https://example.com/x;;Client: type=2\r\n",
			want:  map[string]string{"Cookie": "a=1; b=2", "Referer": "https://example.com/x", "Client": "type=2"},
		},
		{
			name:  "pseudo headers copied from a browser",
			input: ":authority: example.com\n:method: GET\nAccept: */*\nnot a header\n",
			want:  map[string]string{"Accept": "*/*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := input.GetHeaders(tt.input); !maps.Equal(got, tt.want) {
				t.Errorf("GetHeaders(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
