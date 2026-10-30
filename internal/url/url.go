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

package url

import (
	"fmt"
	"net"
	"net/url"
	"path"
	"strings"
	"unicode/utf8"

	errUtils "github.com/edoardottt/cariddi/internal/error"
	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

// Scope defines which URLs can be crawled starting from a target.
type Scope struct {
	host       string // target host or, in intensive mode, its root domain.
	port       string // target port, "" for the default web ports.
	anyPort    bool
	subdomains bool
}

// GetHost takes as input a string and
// tries to parse it as url, if it's a
// well formatted url this function returns
// the host (the domain if you prefer).
func GetHost(input string) string {
	if !HasScheme(input) {
		input = "http://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return ""
	}

	return u.Host
}

// GetProtocol takes as input a string and
// tries to parse it as url, if it's a
// well formatted url this function returns
// the protocol (the scheme if you prefer).
func GetProtocol(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}

	return u.Scheme
}

// GetRootHost takes as input a string and
// tries to parse it as url, if it's a
// well formatted url this function returns
// the root domain: the TLD plus one label
// (e.g. edoardottt.com for a.b.edoardottt.com,
// edoardottt.co.uk for a.b.edoardottt.co.uk).
// IP addresses and single label hosts are
// returned as they are.
func GetRootHost(input string) (string, error) {
	if !HasScheme(input) {
		input = "http://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	host := normalizeHost(u.Hostname())
	if host == "" {
		return "", fmt.Errorf("%w", errUtils.ErrDomainFormat)
	}

	if net.ParseIP(host) != nil {
		return host, nil
	}

	labels := strings.Split(host, ".")

	// The TLD is the longest ICANN public suffix (e.g. com, co.uk):
	// the private ones (e.g. github.io) are ignored.
	for i := range labels {
		suffix := strings.Join(labels[i:], ".")
		if tld, icann := publicsuffix.PublicSuffix(suffix); icann && tld == suffix {
			if i == 0 {
				return "", fmt.Errorf("%w", errUtils.ErrDomainTLD)
			}

			return strings.Join(labels[i-1:], "."), nil
		}
	}

	// Unknown TLD (e.g. internal) or single label host (e.g. localhost).
	if len(labels) > 1 {
		return strings.Join(labels[len(labels)-2:], "."), nil
	}

	return host, nil
}

// NewScope returns the scope of the target: its host and port or,
// in intensive mode, its root domain and all the subdomains on any port.
func NewScope(target string, intensive bool) (*Scope, error) {
	if !HasScheme(target) {
		target = "http://" + target
	}

	u, err := url.Parse(target)
	if err != nil {
		return nil, err
	}

	host := normalizeHost(u.Hostname())
	if host == "" {
		return nil, fmt.Errorf("%w", errUtils.ErrDomainFormat)
	}

	if !intensive {
		return &Scope{host: host, port: webPort(u.Port())}, nil
	}

	root, err := GetRootHost(target)
	if err != nil {
		return nil, err
	}

	return &Scope{host: root, anyPort: true, subdomains: net.ParseIP(root) == nil}, nil
}

// Contains checks if the input is an http(s) URL in scope.
func (s *Scope) Contains(input string) bool {
	u, err := url.Parse(input)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return false
	}

	host := normalizeHost(u.Hostname())
	if host != s.host && (!s.subdomains || !strings.HasSuffix(host, "."+s.host)) {
		return false
	}

	return s.anyPort || webPort(u.Port()) == s.port
}

// normalizeHost lowercases the host, removes the trailing dot and
// converts internationalized domain names to ASCII, as colly does
// for the URLs it visits (e.g. bücher.de -> xn--bcher-kva.de).
func normalizeHost(host string) string {
	host = strings.TrimSuffix(strings.ToLower(host), ".")

	if strings.ContainsFunc(host, func(r rune) bool { return r >= utf8.RuneSelf }) {
		if ascii, err := idna.Lookup.ToASCII(host); err == nil {
			return ascii
		}
	}

	return host
}

// webPort returns the port, or "" for the default web ports (80 and 443):
// http://target.com and https://target.com are the same website.
func webPort(port string) string {
	if port == "80" || port == "443" {
		return ""
	}

	return port
}

// HasProtocol takes as input a string and
// checks if it has a protocol ( like in a
// URI/URL).
func HasProtocol(input string) bool {
	res := strings.Index(input, "://")
	return res >= 0
}

// HasScheme checks if the input starts with a scheme followed by "://"
// (e.g. https://example.com). Unlike HasProtocol, it ignores "://" in
// the rest of the input (e.g. example.com/?next=https://example.org).
func HasScheme(input string) bool {
	scheme, _, found := strings.Cut(input, "://")
	if !found || scheme == "" {
		return false
	}

	// RFC 3986: ALPHA *( ALPHA / DIGIT / "+" / "-" / "." ).
	for i, c := range scheme {
		switch {
		case 'a' <= c && c <= 'z', 'A' <= c && c <= 'Z':
		case i > 0 && ('0' <= c && c <= '9' || c == '+' || c == '-' || c == '.'):
		default:
			return false
		}
	}

	return true
}

// RemovePort removes port from the input string.
// If it's not present it returns the input.
func RemovePort(input string) string {
	res := strings.Index(input, ":")
	if res >= 0 {
		return input[:res]
	}

	return input
}

// RetrieveParameters takes as input a string and
// if it's correctly url-formatted returns a slice
// of strings that are the parameters of the URL.
func RetrieveParameters(input string) []string {
	result := []string{}

	u, err := url.Parse(input)
	if err != nil {
		return result
	}

	m, _ := url.ParseQuery(u.RawQuery)
	for k := range m {
		result = append(result, k)
	}

	return result
}

// AbsoluteURL takes as input a protocol, a domain and a path
// and returns the absolute URL with protocol + domain + path.
func AbsoluteURL(protocol string, target string, path string) string {
	// if the path variable starts with a scheme, it means that the
	// path is itself an absolute path.
	if HasProtocol(path) {
		return path
	}

	if len(path) != 0 && path[0] == '/' {
		return protocol + "://" + target + path
	}

	return protocol + "://" + target + "/" + path
}

// GetPath returns the path of the input string
// (if correctly URL-formatted).
func GetPath(input string) (string, error) {
	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	return u.Path, nil
}

// IsEmailURL checks if the input string is a mail URL.
func IsEmailURL(input string) (bool, string) {
	if input[:7] == "mailto:" {
		return true, input[7:]
	}

	return false, ""
}

// GetURLExtension extracts the file extension (without the dot) from a URL,
// in lowercase. Returns "" if there's no extension.
func GetURLExtension(u *url.URL) string {
	if u == nil {
		return ""
	}

	// u.Path is already decoded: decoding it again would read e.g.
	// /a%252Epng (the file "a%2Epng") as /a.png.
	ext := path.Ext(path.Base(u.Path))

	// Remove the dot and return lowercase
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}
