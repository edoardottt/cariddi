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

package url_test

import (
	"errors"
	"net/url"
	"reflect"
	"testing"

	errUtils "github.com/edoardottt/cariddi/v2/internal/error"
	urlUtils "github.com/edoardottt/cariddi/v2/internal/url"
)

func TestGetHost(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no protocol",
			input: "edoardottt.com/ciao?id=1",
			want:  "edoardottt.com",
		},
		{
			name:  "ok1",
			input: "http://edoardottt.com/ciao?id=1",
			want:  "edoardottt.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlUtils.GetHost(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetProtocol(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no protocol",
			input: "ciao.com",
			want:  "",
		},
		{
			name:  "ok1",
			input: "http://edoardottt.com/ciao?id=1",
			want:  "http",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlUtils.GetProtocol(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRootHost(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no protocol",
			input: "ciao.com",
			want:  "ciao.com",
		},
		{
			name:  "no protocol2",
			input: "ciao.ciao.com",
			want:  "ciao.com",
		},
		{
			name:  "ok1",
			input: "http://sub.edoardottt.com/ciao?id=1",
			want:  "edoardottt.com",
		},
		{
			name:  "nth level subdomain with port",
			input: "https://sub.sub.sub.domain.com:8443/page",
			want:  "domain.com",
		},
		{
			name:  "multi label TLD",
			input: "https://shop.example.co.uk/",
			want:  "example.co.uk",
		},
		{
			name:  "private suffix ignored",
			input: "https://edoardottt.github.io/",
			want:  "github.io",
		},
		{
			name:  "private suffix ignored 2",
			input: "https://github.io/",
			want:  "github.io",
		},
		{
			name:  "uppercase and trailing dot",
			input: "https://SUB.Domain.COM./x",
			want:  "domain.com",
		},
		{
			name:  "internationalized domain name",
			input: "https://shop.BÜCHER.de/",
			want:  "xn--bcher-kva.de",
		},
		{
			name:  "unknown TLD",
			input: "http://app.corp.internal/",
			want:  "corp.internal",
		},
		{
			name:  "single label",
			input: "http://localhost:8080/",
			want:  "localhost",
		},
		{
			name:  "IPv4",
			input: "http://10.0.0.5:8080/",
			want:  "10.0.0.5",
		},
		{
			name:  "IPv6",
			input: "http://[::1]:8080/",
			want:  "::1",
		},
		{
			name:  "TLD",
			input: "https://co.uk/",
			want:  "",
		},
		{
			name:  "TLD 2",
			input: "com",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := urlUtils.GetRootHost(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRootHost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewScopeErrors(t *testing.T) {
	tests := []struct {
		target    string
		intensive bool
		want      error
	}{
		{target: "", intensive: false, want: errUtils.ErrDomainFormat},
		{target: "https://co.uk/", intensive: true, want: errUtils.ErrDomainTLD},
		{target: "com", intensive: true, want: errUtils.ErrDomainTLD},
		{target: "https://co.uk/", intensive: false, want: nil},
		{target: "https://github.io/", intensive: true, want: nil},
	}
	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			if _, err := urlUtils.NewScope(tt.target, tt.intensive); !errors.Is(err, tt.want) {
				t.Errorf("NewScope(%q, %v) error = %v, want %v", tt.target, tt.intensive, err, tt.want)
			}
		})
	}
}

func TestScopeContains(t *testing.T) {
	tests := []struct {
		target    string
		intensive bool
		input     string
		want      bool
	}{
		{target: "https://sub.domain.com:8443/page", input: "https://sub.domain.com:8443/other", want: true},
		{target: "https://sub.domain.com:8443/page", input: "http://sub.domain.com:8443/", want: true},
		{target: "https://sub.domain.com:8443/page", input: "https://sub.domain.com/", want: false},
		{target: "https://sub.domain.com:8443/page", input: "https://x.sub.domain.com:8443/", want: false},
		{target: "https://sub.domain.com:8443/page", input: "https://domain.com:8443/", want: false},
		{target: "https://sub.domain.com/page", input: "http://sub.domain.com/", want: true},
		{target: "https://sub.domain.com/page", input: "https://sub.domain.com:443/", want: true},
		{target: "https://sub.domain.com/page", input: "https://SUB.domain.com./", want: true},
		{target: "https://sub.domain.com/page", input: "https://sub.domain.com:8443/", want: false},
		{target: "https://sub.domain.com/page", input: "ftp://sub.domain.com/", want: false},
		{target: "https://sub.domain.com/page", input: "mailto:a@sub.domain.com", want: false},
		{target: "http://bücher.de/", input: "http://xn--bcher-kva.de/x", want: true},
		{target: "http://xn--bcher-kva.de/", input: "http://BÜCHER.de/x", want: true},
		{target: "http://[::1]:8080/", input: "http://[::1]:8080/x", want: true},
		{target: "http://[::1]:8080/", input: "http://[::1]:9090/", want: false},
		{target: "https://sub.sub.sub.domain.com:8443/page", intensive: true, input: "https://domain.com/", want: true},
		{target: "https://sub.sub.sub.domain.com:8443/page", intensive: true, input: "https://sub.domain.com/", want: true},
		{target: "https://sub.sub.sub.domain.com:8443/page", intensive: true, input: "http://a.b.c.domain.com:8080/", want: true},
		{target: "https://sub.sub.sub.domain.com:8443/page", intensive: true, input: "https://fakedomain.com/", want: false},
		{target: "https://sub.sub.sub.domain.com:8443/page", intensive: true, input: "https://x.fakedomain.com/", want: false},
		{target: "https://sub.sub.sub.domain.com:8443/page", intensive: true, input: "https://domain.com.evil.com/", want: false},
		{target: "https://shop.example.co.uk/", intensive: true, input: "https://www.example.co.uk/", want: true},
		{target: "https://shop.example.co.uk/", intensive: true, input: "https://other.co.uk/", want: false},
		{target: "https://shop.bücher.de/", intensive: true, input: "https://www.xn--bcher-kva.de/", want: true},
		{target: "https://edoardottt.github.io/", intensive: true, input: "https://other.github.io/", want: true},
		{target: "https://edoardottt.github.io/", intensive: true, input: "https://fakegithub.io/", want: false},
		{target: "https://edoardottt.github.io/", intensive: true, input: "https://subdomain.fakegithub.io/", want: false},
		{target: "http://10.0.0.5:8080/", intensive: true, input: "http://10.0.0.5:9090/", want: true},
		{target: "http://10.0.0.5:8080/", intensive: true, input: "http://10.0.0.50/", want: false},
		{target: "http://10.0.0.5:8080/", intensive: true, input: "http://x.10.0.0.5/", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.target+" "+tt.input, func(t *testing.T) {
			scope, err := urlUtils.NewScope(tt.target, tt.intensive)
			if err != nil {
				t.Fatal(err)
			}

			if got := scope.Contains(tt.input); got != tt.want {
				t.Errorf("scope of %s (intensive: %v) contains %s = %v, want %v",
					tt.target, tt.intensive, tt.input, got, tt.want)
			}
		})
	}
}

func TestHasProtocol(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{
			name:  "empty string",
			input: "",
			want:  false,
		},
		{
			name:  "no protocol",
			input: "ciao.com",
			want:  false,
		},
		{
			name:  "no protocol2",
			input: "ciao.ciao.com",
			want:  false,
		},
		{
			name:  "ok1",
			input: "http://sub.edoardottt.com/ciao?id=1",
			want:  true,
		},
		{
			name:  "ok2",
			input: "ftp://sub.edoardottt.com/://ciao",
			want:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlUtils.HasProtocol(tt.input); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HasProtocol() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetURLExtension(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"https://example.com/a.png", "png"},
		{"https://example.com/A.PNG", "png"},
		{"https://example.com/dir/a.min.js?v=1#top", "js"},
		{"https://example.com/a%2Epng", "png"},    // %2E is a dot.
		{"https://example.com/a%252Epng", ""},     // the file a%2Epng.
		{"https://example.com/100%25.png", "png"}, // the file 100%.png.
		{"https://example.com/dir.png/file", ""},
		{"https://example.com/a.png%5Cb", `png\b`}, // the same on every OS.
		{"https://example.com/", ""},
	}
	for _, tt := range tests {
		u, err := url.Parse(tt.input)
		if err != nil {
			t.Fatal(err)
		}

		if got := urlUtils.GetURLExtension(u); got != tt.want {
			t.Errorf("GetURLExtension(%s) = %q, want %q", tt.input, got, tt.want)
		}
	}

	if got := urlUtils.GetURLExtension(nil); got != "" {
		t.Errorf("GetURLExtension(nil) = %q, want \"\"", got)
	}
}
