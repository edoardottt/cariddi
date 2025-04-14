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

	@Author:      edoardottt, https://www.edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package url

import (
	"fmt"
	"net/url"
	"path"
	"path/filepath"
	"strings"

	errUtils "github.com/edoardottt/cariddi/internal/error"
)

// GetHost takes as input a string and
// tries to parse it as url, if it's a
// well formatted url this function returns
// the host (the domain if you prefer).
func GetHost(input string) string {
	if !HasProtocol(input) {
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
// the second level domain.
func GetRootHost(input string) (string, error) {
	if !HasProtocol(input) {
		input = "http://" + input
	}

	u, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	// divide host and port, then split by dot
	parts := strings.Split(strings.Split(u.Host, ":")[0], ".")
	// return the last two parts
	if len(parts) > 1 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1], nil
	}

	return "", fmt.Errorf("%w", errUtils.ErrDomainFormat)
}

// HasProtocol takes as input a string and
// checks if it has a protocol ( like in a
// URI/URL).
func HasProtocol(input string) bool {
	res := strings.Index(input, "://")
	return res >= 0
}

// RemoveProtocol removes the protocol from
// the input string (something://...).
// If it's not present it returns the input.
func RemoveProtocol(input string) string {
	res := strings.Index(input, "://")
	if res >= 0 {
		return input[res+3:]
	}

	return input
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

// SameDomain checks if two urls have the same domain.
func SameDomain(url1 string, url2 string) bool {
	u1, err := url.Parse(url1)
	if err != nil {
		return false
	}

	u2, err := url.Parse(url2)
	if err != nil {
		return false
	}

	if u1.Host == "" || u2.Host == "" {
		return false
	}

	return u1.Host == u2.Host
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

// GetURLExtension extracts the file extension (without the dot) from a URL.
// It decodes the URL path first and returns the extension in lowercase.
// Returns "" if there's no extension.
func GetURLExtension(u *url.URL) string {
	if u == nil {
		return ""
	}

	// Decode percent-encoded characters in the path
	decodedPath, err := url.PathUnescape(u.Path)
	if err != nil {
		return ""
	}

	// Get the last part of the path (e.g. file name)
	lastPart := path.Base(decodedPath)

	// Get extension (includes the dot, e.g. ".pdf")
	ext := filepath.Ext(lastPart)
	if ext == "" {
		return ""
	}

	// Remove the dot and return lowercase
	return strings.ToLower(strings.TrimPrefix(ext, "."))
}
