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
	"fmt"
	"math"
	"math/big"
	"net/url"
	"reflect"
	"slices"
	"strings"
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

func TestParseSrcset(t *testing.T) {
	// The numbers halfway between the largest float64 and 2^1024 are rounded
	// to 2^1024 (an overflow), as the spec treats it as even.
	maxFloat, _ := new(big.Float).SetFloat64(math.MaxFloat64).Int(nil)
	prevFloat, _ := new(big.Float).SetFloat64(math.Nextafter(math.MaxFloat64, 0)).Int(nil)
	halfway := new(big.Int).Add(maxFloat, new(big.Int).Rsh(new(big.Int).Sub(maxFloat, prevFloat), 1))
	underHalfway := new(big.Int).Sub(halfway, big.NewInt(1))

	tests := []struct {
		name   string
		srcset string
		want   []string
	}{
		// Separators.
		{name: "empty", srcset: "", want: nil},
		{name: "whitespace only", srcset: " \t\n\f\r ", want: nil},
		{name: "commas only", srcset: " , ,, ", want: nil},
		{name: "one URL", srcset: "a.jpg", want: []string{"a.jpg"}},
		{name: "densities", srcset: "a.jpg 1x, b.jpg 2x", want: []string{"a.jpg", "b.jpg"}},
		{name: "widths", srcset: "a.jpg 400w, b.jpg 800w", want: []string{"a.jpg", "b.jpg"}},
		{name: "no space after the comma", srcset: "a.jpg 1x,b.jpg 2x", want: []string{"a.jpg", "b.jpg"}},
		{name: "space before the comma", srcset: "a.jpg 1x , b.jpg 2x", want: []string{"a.jpg", "b.jpg"}},
		{name: "comma ending the URL", srcset: "a.jpg, b.jpg 2x", want: []string{"a.jpg", "b.jpg"}},
		{name: "commas ending the URL", srcset: "a.jpg,,, b.jpg", want: []string{"a.jpg", "b.jpg"}},
		{name: "commas between candidates", srcset: "a.jpg 1x,, ,b.jpg 2x", want: []string{"a.jpg", "b.jpg"}},
		{name: "leading and trailing commas", srcset: " ,a.jpg 1x, b.jpg 2x, ", want: []string{"a.jpg", "b.jpg"}},
		{
			name:   "one candidate per line",
			srcset: "\n\t\ta.jpg\t400w,\r\n\t\tb.jpg\f800w\r\n\t",
			want:   []string{"a.jpg", "b.jpg"},
		},
		{name: "fullwidth comma", srcset: "a.jpg 1x\uff0c b.jpg 2x", want: nil},
		{name: "missing comma", srcset: "a.jpg 1x b.jpg 2x", want: nil},
		{name: "comma in a URL", srcset: "a.jpg,b.jpg 2x", want: []string{"a.jpg,b.jpg"}},
		{name: "comma in a URL 2", srcset: "a.jpg,1x", want: []string{"a.jpg,1x"}},

		// Commas in the URLs.
		{
			name: "path transformations",
			srcset: "https://res.cloudinary.com/demo/image/upload/w_400,h_300,c_fill/sample.jpg 400w, " +
				"https://res.cloudinary.com/demo/image/upload/w_800,h_600,c_fill/sample.jpg 800w",
			want: []string{
				"https://res.cloudinary.com/demo/image/upload/w_400,h_300,c_fill/sample.jpg",
				"https://res.cloudinary.com/demo/image/upload/w_800,h_600,c_fill/sample.jpg",
			},
		},
		{
			name:   "path transformations 2",
			srcset: "/cdn-cgi/image/width=400,quality=75/uploads/hero.jpg 400w",
			want:   []string{"/cdn-cgi/image/width=400,quality=75/uploads/hero.jpg"},
		},
		{
			name:   "path transformations 3",
			srcset: "https://ik.imagekit.io/demo/tr:w-300,h-300/cafe.jpg 300w",
			want:   []string{"https://ik.imagekit.io/demo/tr:w-300,h-300/cafe.jpg"},
		},
		{
			name:   "path filters with parentheses",
			srcset: "https://thumbor.example.com/unsafe/300x200/filters:fill(white,1):quality(80)/a.jpg 1x",
			want:   []string{"https://thumbor.example.com/unsafe/300x200/filters:fill(white,1):quality(80)/a.jpg"},
		},
		{
			name: "path coordinates",
			srcset: "https://api.mapbox.com/styles/v1/mapbox/streets-v12/static/-122.42,37.78,14/600x400?access_token=pk.x 1x," +
				"https://api.mapbox.com/styles/v1/mapbox/streets-v12/static/-122.42,37.78,14/600x400@2x?access_token=pk.x 2x",
			want: []string{
				"https://api.mapbox.com/styles/v1/mapbox/streets-v12/static/-122.42,37.78,14/600x400?access_token=pk.x",
				"https://api.mapbox.com/styles/v1/mapbox/streets-v12/static/-122.42,37.78,14/600x400@2x?access_token=pk.x",
			},
		},
		{
			name:   "query list",
			srcset: "https://assets.imgix.net/photo.jpg?auto=format,compress&w=400 400w",
			want:   []string{"https://assets.imgix.net/photo.jpg?auto=format,compress&w=400"},
		},
		{
			name:   "query rectangle",
			srcset: "/images/photo.jpg?rect=0,0,800,600&w=400 400w, /images/photo.jpg?rect=0,0,800,600&w=800 800w",
			want:   []string{"/images/photo.jpg?rect=0,0,800,600&w=400", "/images/photo.jpg?rect=0,0,800,600&w=800"},
		},
		{
			name:   "query size",
			srcset: "https://i0.wp.com/example.com/a.jpg?resize=300,200&ssl=1 300w",
			want:   []string{"https://i0.wp.com/example.com/a.jpg?resize=300,200&ssl=1"},
		},
		{
			name:   "encoded comma",
			srcset: "https://i0.wp.com/example.com/a.jpg?resize=300%2C200&ssl=1 300w",
			want:   []string{"https://i0.wp.com/example.com/a.jpg?resize=300%2C200&ssl=1"},
		},
		{
			name:   "consecutive commas in a URL",
			srcset: "/a,,b.jpg 1x",
			want:   []string{"/a,,b.jpg"},
		},
		{
			name:   "data URI",
			srcset: "data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7",
			want:   []string{"data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7"},
		},
		{
			name:   "encoded SVG data URI",
			srcset: "data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2016%209'%3E%3C/svg%3E 1x",
			want:   []string{"data:image/svg+xml,%3Csvg%20xmlns='http://www.w3.org/2000/svg'%20viewBox='0%200%2016%209'%3E%3C/svg%3E"},
		},
		{
			name:   "SVG data URI with spaces",
			srcset: `data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg"></svg>`,
			want:   nil,
		},

		// Other URLs.
		{name: "parentheses in a URL", srcset: "/a(1).jpg 1x, (b).jpg 2x", want: []string{"/a(1).jpg", "(b).jpg"}},
		{name: "URL like a descriptor", srcset: "2x 2x, 400w", want: []string{"2x", "400w"}},
		{
			name:   "relative URLs",
			srcset: "../a.jpg 1x, //cdn.example.com/b.jpg 2x, ?v=3 3x, #x 4x",
			want:   []string{"../a.jpg", "//cdn.example.com/b.jpg", "?v=3", "#x"},
		},
		{name: "non-ASCII URL", srcset: "/città/图片,1.jpg 1x", want: []string{"/città/图片,1.jpg"}},
		{
			name:   "non-ASCII whitespace in a URL",
			srcset: "/a\u00a0b\vc\u0085d\u3000e.jpg 1x",
			want:   []string{"/a\u00a0b\vc\u0085d\u3000e.jpg"},
		},
		{name: "invalid UTF-8", srcset: "/\xff\xfe.jpg 1x", want: []string{"/\xff\xfe.jpg"}},
		{name: "template", srcset: "{{ image.src }} 1x", want: nil},
		{name: "template 2", srcset: "{{image.src}} 1x", want: []string{"{{image.src}}"}},
		{name: "duplicate URLs", srcset: "a.jpg 1x, a.jpg 2x", want: []string{"a.jpg", "a.jpg"}},
		{name: "duplicate densities", srcset: "a.jpg, b.jpg 1x", want: []string{"a.jpg", "b.jpg"}},

		// Valid descriptors.
		{name: "density", srcset: "a.jpg 1.5x", want: []string{"a.jpg"}},
		{name: "zero density", srcset: "a.jpg 0x", want: []string{"a.jpg"}},
		{name: "negative zero density", srcset: "a.jpg -0.0e5x", want: []string{"a.jpg"}},
		{name: "density underflow", srcset: "a.jpg 1e-400x", want: []string{"a.jpg"}},
		{name: "negative density underflow", srcset: "a.jpg -1e-400x", want: []string{"a.jpg"}},
		{name: "huge exponent of zero", srcset: "a.jpg 0e99999999999999999999x", want: []string{"a.jpg"}},
		{name: "huge negative exponent", srcset: "a.jpg 1e-99999999999999999999x", want: []string{"a.jpg"}},
		{name: "max density", srcset: "a.jpg 1.7976931348623157e308x", want: []string{"a.jpg"}},
		{name: "density under the overflow", srcset: "a.jpg " + underHalfway.String() + "x", want: []string{"a.jpg"}},
		{name: "exponent with plus", srcset: "a.jpg 1e+1x", want: []string{"a.jpg"}},
		{name: "leading zeros", srcset: "a.jpg 007w, b.jpg 00.50x", want: []string{"a.jpg", "b.jpg"}},
		{name: "huge width", srcset: "a.jpg 99999999999999999999999999w", want: []string{"a.jpg"}},
		{name: "width and height", srcset: "a.jpg 400w 300h, b.jpg 600h 800w", want: []string{"a.jpg", "b.jpg"}},
		{name: "whitespace around descriptors", srcset: "a.jpg \t 400w \n 300h \f, b.jpg", want: []string{"a.jpg", "b.jpg"}},

		// Invalid descriptors.
		{name: "uppercase", srcset: "a.jpg 2X, b.jpg 400W", want: nil},
		{name: "unit", srcset: "a.jpg 400px", want: nil},
		{name: "two densities", srcset: "a.jpg 1x 2x", want: nil},
		{name: "width and density", srcset: "a.jpg 400w 1x, b.jpg 1x 400w", want: nil},
		{name: "height only", srcset: "a.jpg 300h", want: nil},
		{name: "two heights", srcset: "a.jpg 400w 300h 300h", want: nil},
		{name: "zero width", srcset: "a.jpg 000w", want: nil},
		{name: "density overflow", srcset: "a.jpg 1e400x", want: nil},
		{name: "density overflow, halfway", srcset: "a.jpg " + halfway.String() + "x", want: nil},
		{name: "huge exponent", srcset: "a.jpg 1e99999999999999999999x", want: nil},
		{name: "no exponent digits", srcset: "a.jpg 1ex, b.jpg 1e+x, c.jpg 1e-+1x", want: nil},
		{name: "two exponents", srcset: "a.jpg 1e1e1x", want: nil},
		{name: "two decimal points", srcset: "a.jpg 1.2.3x", want: nil},
		{name: "two minus signs", srcset: "a.jpg --1x", want: nil},
		{name: "descriptors without separator", srcset: "a.jpg 400w300h", want: nil},
		{name: "unit only", srcset: "a.jpg w, b.jpg x, c.jpg h", want: nil},
		{name: "future descriptor", srcset: "a.jpg (1, 2), b.jpg 2x", want: []string{"b.jpg"}},
		{name: "unclosed parenthesis", srcset: "a.jpg 1x, b.jpg (, c.jpg 2x", want: []string{"a.jpg"}},
		{
			name:   "invalid candidate between valid ones",
			srcset: "a.jpg 1x, b.jpg 2X, c.jpg 3x",
			want:   []string{"a.jpg", "c.jpg"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlUtils.ParseSrcset(tt.srcset); !slices.Equal(got, tt.want) {
				t.Errorf("ParseSrcset(%q) = %q, want %q", tt.srcset, got, tt.want)
			}
		})
	}
}

// TestParseSrcsetWPT checks the web-platform-tests cases of the srcset parsing
// (html/semantics/embedded-content/the-img-element/srcset/parse-a-srcset-attribute.html),
// with the URLs of all the candidates instead of the one selected.
func TestParseSrcsetWPT(t *testing.T) {
	check := func(srcset string, want []string) {
		t.Helper()

		if got := urlUtils.ParseSrcset(srcset); !slices.Equal(got, want) {
			t.Errorf("ParseSrcset(%q) = %q, want %q", srcset, got, want)
		}
	}

	a := []string{"data:,a"}

	for _, tt := range []struct {
		srcset string
		want   []string
	}{
		// Splitting loop.
		{"", nil},
		{",", nil},
		{",,,", nil},
		{"  data:,a  1x  ", a},
		{"\t\tdata:,a\t\t1x\t\t", a},
		{"\n\ndata:,a\n\n1x\n\n", a},
		{"\v\vdata:,a\v\v1x\v\v", []string{"\v\vdata:,a\v\v1x\v\v"}},
		{"\f\fdata:,a\f\f1x\f\f", a},
		{"\r\rdata:,a\r\r1x\r\r", a},
		{"\x0e\x0edata:,a\x0e\x0e1x\x0e\x0e", []string{"\x0e\x0edata:,a\x0e\x0e1x\x0e\x0e"}},
		{"\x0f\x0fdata:,a\x0f\x0f1x\x0f\x0f", []string{"\x0f\x0fdata:,a\x0f\x0f1x\x0f\x0f"}},
		{"\x10\x10data:,a\x10\x101x\x10\x10", []string{"\x10\x10data:,a\x10\x101x\x10\x10"}},
		{"data:,a", a},
		{"data:,a ", a},
		{"data:,a ,", a},
		{"data:,a,", a},
		{"data:,a, ", a},
		{"data:,a,,,", a},
		{"data:,a,, , ", a},
		{" data:,a", a},
		{",,,data:,a", a},
		{" , ,,data:,a", a},
		{"\u00a0data:,a", []string{"\u00a0data:,a"}},
		{"data:,a\u00a0", []string{"data:,a\u00a0"}},

		// Descriptor tokenizer.
		{"data:,a 1x", a},
		{"data:,a 1x ", a},
		{"data:,a 1x,", a},
		{"data:,a ( , data:,b 1x, ), data:,c", []string{"data:,c"}},
		{"data:,a ((( , data:,b 1x, ), data:,c", []string{"data:,c"}},
		{"data:,a [ , data:,b 1x, ], data:,c", []string{"data:,b", "]", "data:,c"}},
		{"data:,a { , data:,b 1x, }, data:,c", []string{"data:,b", "}", "data:,c"}},
		{`data:,a " , data:,b 1x, ", data:,c`, []string{"data:,b", `"`, "data:,c"}},
		{`data:,a \,data:;\,b, data:,c`, []string{`data:;\,b`, "data:,c"}},
		{"data:,a, data:,b (", a},
		{"data:,a, data:,b (  ", a},
		{"data:,a, data:,b (,", a},
		{"data:,a, data:,b (x", a},
		{"data:,a, data:,b ()", a},
		{"data:,a (, data:,b", nil},
		{"data:,a /*, data:,b, data:,c */", []string{"data:,b"}},
		{"data:,a //, data:,b", []string{"data:,b"}},
	} {
		check(tt.srcset, tt.want)
	}

	// Descriptor parser.
	for _, d := range []string{
		"1w 1h", "1h 1w", "1w", "0x", "-0x", "1e0x", "1E0x", "1e-1x", "1.5e1x", ".5x", ".5e1x", "1.0x",
	} {
		check("data:,a "+d, a)
	}

	for _, d := range []string{
		"foo", "foo foo", "foo 1x", "foo 1x foo", "foo 1w", "foo 1w foo", "1x 1x", "1w 1w", "1w 1x", "1x 1w",
		"1h 1h", "1h 1x", "1h 1w 1x", "1x 1w 1h", "1h", "1h foo", "foo 1h", "0w", "-1w", "1w -1w", "1.0w",
		"1w 1.0w", "1e0w", "1w 1e0w", "1www", "1w 1www", "+1w", "1w +1w", "1W", "1w 1W", "Infinityw",
		"1w Infinityw", "NaNw", "1w NaNw", "0x1w", "0X1w", "1x -0x", "-1x", "1x -1x", "-x", ".x", "-.x",
		"1.x", "1x 1.5e1x", "1x 1e1.5x", "1x 1.0x", "+1x", "1X", "Infinityx", "NaNx", "0x1x", "0X1x",
		"1w 0h", "1w -1h", "1w 1.0h", "1w 1e0h", "1w 1hhh", "1w +1h", "1w 1H", "1w Infinityh", "1w NaNh",
		"0x1h", "0X1h",
	} {
		check("data:,a "+d, nil)
	}

	// Any other character is part of a descriptor, even the other
	// whitespace of unicode.IsSpace (U+000B, U+0085, U+2028, U+2029).
	for _, c := range "\x01\u00a0\u1680\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200a" +
		"\u200c\u200d\u202f\u205f\u3000\ufeff\v\u0085\u2028\u2029" {
		for _, format := range []string{
			"data:,a 1%cw", "data:,a %c1w", "data:,a 1%cx", "data:,a %c1x", "data:,a 1w 1%ch", "data:,a 1w %c1h",
		} {
			check(fmt.Sprintf(format, c), nil)
		}
	}
}

// TestParseSrcsetRoundTrip checks that two valid candidates, with any URL,
// descriptors and separator, and an invalid candidate or not between them,
// are parsed back to their URLs.
func TestParseSrcsetRoundTrip(t *testing.T) {
	urls := []string{
		"a.jpg",
		"2x",
		"/a(1).jpg",
		"/a,,b.jpg",
		"https://res.cloudinary.com/demo/image/upload/w_400,h_300,c_fill/sample.jpg",
		"/cdn-cgi/image/width=400,quality=75/uploads/hero.jpg",
		"https://assets.imgix.net/photo.jpg?auto=format,compress&w=400",
		"https://thumbor.example.com/unsafe/filters:fill(white,1)/a.jpg",
		"data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7",
		"/città/图片,1.jpg",
		"/a\u00a0b\vc.jpg",
	}
	descriptors := []string{"", " ", " 1x", "\t1.5x\n", " 400w", "  400w   300h  ", "\r\n300h 400w"}
	separators := []string{",", ",,", " ,", ", ", " , ", ",\n\t", ", ,"}
	invalid := []string{"", "b.jpg 1x 2x, ", "c.jpg (1, 2), "}

	for _, u1 := range urls {
		for _, d1 := range descriptors {
			for _, sep := range separators {
				// A URL followed by a comma only is a URL with a comma.
				if d1 == "" && strings.Trim(sep, ",") == "" {
					continue
				}

				for _, between := range invalid {
					for _, u2 := range urls {
						for _, d2 := range descriptors {
							srcset := u1 + d1 + sep + between + u2 + d2
							if got := urlUtils.ParseSrcset(srcset); !slices.Equal(got, []string{u1, u2}) {
								t.Fatalf("ParseSrcset(%q) = %q, want %q", srcset, got, []string{u1, u2})
							}
						}
					}
				}
			}
		}
	}
}

// FuzzParseSrcset checks that the URLs are in srcset, in order, without
// whitespace or commas around them, and that a srcset of these URLs only
// is parsed back to them.
func FuzzParseSrcset(f *testing.F) {
	for _, seed := range []string{
		"a.jpg 1x, b.jpg 2x",
		"https://res.cloudinary.com/demo/image/upload/w_400,h_300,c_fill/sample.jpg 400w,a.jpg,, b.jpg",
		"data:,a ( , data:,b 1x, ), data:,c",
		`data:,a \,data:;\,b, data:,c`,
		"a.jpg 400w 300h, b.jpg 1e-400x, c.jpg (",
		"\v\u00a0a.jpg\f1x\r\n,",
	} {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, srcset string) {
		urls := urlUtils.ParseSrcset(srcset)

		rest := srcset
		for _, u := range urls {
			i := strings.Index(rest, u)
			if u == "" || i < 0 || strings.ContainsAny(u, " \t\n\f\r") || u[0] == ',' || u[len(u)-1] == ',' {
				t.Fatalf("ParseSrcset(%q) = %q: bad URL %q", srcset, urls, u)
			}

			rest = rest[i+len(u):]
		}

		joined := strings.Join(urls, ", ")
		if got := urlUtils.ParseSrcset(joined); !slices.Equal(got, urls) {
			t.Fatalf("ParseSrcset(%q) = %q, want %q", joined, got, urls)
		}
	})
}

// BenchmarkParseSrcset measures the parsing of a srcset with commas in the URLs.
func BenchmarkParseSrcset(b *testing.B) {
	srcset := "https://res.cloudinary.com/demo/image/upload/w_400,h_300,c_fill/sample.jpg 400w, " +
		"https://res.cloudinary.com/demo/image/upload/w_800,h_600,c_fill/sample.jpg 800w, " +
		"https://res.cloudinary.com/demo/image/upload/w_1200,h_900,c_fill/sample.jpg 1200w, " +
		"https://res.cloudinary.com/demo/image/upload/w_1600,h_1200,c_fill/sample.jpg 1600w"

	b.SetBytes(int64(len(srcset)))

	for b.Loop() {
		urlUtils.ParseSrcset(srcset)
	}
}
