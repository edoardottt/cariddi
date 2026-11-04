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
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/edoardottt/cariddi/v2/pkg/input"
)

func TestParseHeaders(t *testing.T) {
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
		{
			name:  "no header",
			input: "headers.txt",
			want:  map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := input.ParseHeaders(tt.input); !maps.Equal(got, tt.want) {
				t.Errorf("ParseHeaders(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// writeFile writes content into a new file of the test, returning its path.
func writeFile(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "values.txt")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

func TestGetHeaders(t *testing.T) {
	file := writeFile(t, "Cookie: from=file\nAuthorization: Bearer abc#def\n")

	// A file and values, in order: the last value of a header wins.
	// A blank value adds nothing.
	got := input.GetHeaders([]string{"Cookie: first;;Client: type=2", file, " ", "Client: type=3"})
	want := map[string]string{"Cookie": "from=file", "Authorization": "Bearer abc#def", "Client": "type=3"}

	if !maps.Equal(got, want) {
		t.Errorf("GetHeaders = %v, want %v", got, want)
	}
}

func TestGetIgnore(t *testing.T) {
	file := writeFile(t, "  logout  \n\n#/admin\nforum, blog\n")
	dir := t.TempDir()

	// A directory is not read: it's a string to ignore.
	got := input.GetIgnore([]string{"delete, logout,,", file, "", dir})
	want := []string{"delete", "logout", "#/admin", "forum, blog", dir}

	if !slices.Equal(got, want) {
		t.Errorf("GetIgnore = %q, want %q", got, want)
	}
}

func TestReadTargets(t *testing.T) {
	// With its line break, this line fills the read buffer.
	limit := strings.Repeat("a", input.MaxTargetLength-1)

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "spaces, blank lines, crlf and no final line break",
			input: "  https://a.com/?q=A B  \r\n\n\r\nhttps://b.com",
			want:  []string{"https://a.com/?q=A B", "https://b.com"},
		},
		{
			name:  "line at the limit",
			input: limit + "\nhttps://b.com\n",
			want:  []string{limit, "https://b.com"},
		},
		{
			name:  "line over the limit",
			input: limit + "a\nhttps://b.com\n",
			want:  []string{"https://b.com"},
		},
		{
			// Its end is skipped too.
			name:  "line longer than the buffer many times",
			input: "https://a.com\n" + strings.Repeat("a", 3*input.MaxTargetLength+5) + "\r\nhttps://b.com",
			want:  []string{"https://a.com", "https://b.com"},
		},
		{
			name:  "last line over the limit, without line break",
			input: "https://a.com\n" + limit + "a",
			want:  []string{"https://a.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := input.ReadTargets(strings.NewReader(tt.input)); !slices.Equal(got, tt.want) {
				t.Errorf("ReadTargets = %d targets, want %d", len(got), len(tt.want))
			}
		})
	}
}

func TestHelpRequested(t *testing.T) {
	tests := []struct {
		args []string
		want bool
	}{
		{args: nil, want: false},
		{args: []string{"-h"}, want: true},
		{args: []string{"--help"}, want: true},
		{args: []string{"-h", "hunt"}, want: true},
		{args: []string{"-s", "-d", "2", "-h"}, want: true},
		// -h and help are the values of -u and -i.
		{args: []string{"-u", "-h"}, want: false},
		{args: []string{"-i", "help"}, want: false},
		// The parsing stops at the first argument that is not a flag.
		{args: []string{"https://example.com", "-h"}, want: false},
		{args: []string{"--", "-h"}, want: false},
		// A parse error prints the help only if an argument asks for it.
		{args: []string{"-x"}, want: false},
		{args: []string{"-x", "-h"}, want: true},
		{args: []string{"-d", "abc", "https://example.com", "-h"}, want: true},
		{args: []string{"-u", "-h", "-d"}, want: true},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			if got := input.HelpRequested(tt.args); got != tt.want {
				t.Errorf("HelpRequested(%q) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestScanFlag(t *testing.T) {
	defaults := strings.Split(input.DefaultIgnoreExtensions, ",")

	tests := []struct {
		name  string
		args  []string
		check func(flags input.Input, unexpected []string) bool
	}{
		{
			name: "defaults",
			args: []string{"-silent"},
			check: func(f input.Input, unexpected []string) bool {
				return f.Threads == input.DefaultThreads && f.Timeout == input.TimeoutRequest &&
					f.MaxDepth == 0 && !f.Secrets && !f.Endpoints && slices.Equal(f.IgnoreExtensions, defaults) &&
					len(unexpected) == 0
			},
		},
		{
			// Commas are part of a URL, and -u is repeatable.
			name: "urls",
			args: []string{"-u", "https://a.com/?ids=1,2", "-url", "https://b.com"},
			check: func(f input.Input, _ []string) bool {
				return slices.Equal(f.URLs, []string{"https://a.com/?ids=1,2", "https://b.com"})
			},
		},
		{
			// The values are kept as they are: GetHeaders and GetIgnore split them and read the files.
			name: "headers and ignore",
			args: []string{"-H", "Authorization: Bearer abc#def", "-headers", "Accept: a,b", "-i", "a, b", "-ignore", "#x"},
			check: func(f input.Input, _ []string) bool {
				return slices.Equal(f.Headers, []string{"Authorization: Bearer abc#def", "Accept: a,b"}) &&
					slices.Equal(f.Ignore, []string{"a, b", "#x"})
			},
		},
		{
			name: "ignore extensions replace the default ones",
			args: []string{"-ie", "GIF, .Svg", "-ignore-ext", "png"},
			check: func(f input.Input, _ []string) bool {
				return slices.Equal(f.IgnoreExtensions, []string{"gif", ".svg", "png"})
			},
		},
		{
			name: "no ignored extension",
			args: []string{"-ie", ""},
			check: func(f input.Input, _ []string) bool {
				return len(f.IgnoreExtensions) == 0
			},
		},
		{
			name: "custom files enable their hunt",
			args: []string{"-sf", "secrets.txt", "-endpoints-file", "endpoints.txt"},
			check: func(f input.Input, _ []string) bool {
				return f.Secrets && f.Endpoints && f.SecretsFile == "secrets.txt" && f.EndpointsFile == "endpoints.txt"
			},
		},
		{
			// -d is the depth, -t the threads: not the delay and the timeout.
			// -l is repeatable, and a file name can contain commas.
			name: "short and long names",
			args: []string{
				"-d", "2", "-t", "5", "-rd", "1", "-timeout", "3", "-subs", "-j", "-v",
				"-l", "urls.txt", "-list", "more,urls.txt",
			},
			check: func(f input.Input, _ []string) bool {
				return f.MaxDepth == 2 && f.Threads == 5 && f.Delay == 1 && f.Timeout == 3 &&
					f.Subdomains && f.JSON && f.Version && slices.Equal(f.List, []string{"urls.txt", "more,urls.txt"})
			},
		},
		{
			// The flags after the first argument that is not a flag are not parsed.
			name: "unexpected arguments",
			args: []string{"https://a.com", "-s"},
			check: func(f input.Input, unexpected []string) bool {
				return !f.Secrets && slices.Equal(unexpected, []string{"https://a.com", "-s"})
			},
		},
		{
			name: "unexpected argument after a URL",
			args: []string{"-u", "https://a.com", "https://b.com"},
			check: func(f input.Input, unexpected []string) bool {
				return slices.Equal(f.URLs, []string{"https://a.com"}) && slices.Equal(unexpected, []string{"https://b.com"})
			},
		},
		{
			name: "end of the flags",
			args: []string{"-s", "--"},
			check: func(f input.Input, unexpected []string) bool {
				return f.Secrets && len(unexpected) == 0
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if flags, unexpected := input.ScanFlagArgs(tt.args); !tt.check(flags, unexpected) {
				t.Errorf("ScanFlag(%q) = %+v, unexpected %q", tt.args, flags, unexpected)
			}
		})
	}
}
