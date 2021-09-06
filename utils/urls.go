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

	@Author:      edoardottt, https://www.edoardoottavianelli.it
*/

package utils

import (
	"net/url"
	"strings"

	"github.com/bobesa/go-domain-util/domainutil"
)

//GetHost >
func GetHost(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	return u.Host
}

//GetProtocol >
func GetProtocol(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	return u.Scheme
}

//GetRootHost >
func GetRootHost(input string) string {
	_, err := url.Parse(input)
	if err != nil {
		return ""
	}
	return domainutil.Domain(input)
}

//GetScheme >
func GetScheme(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	return u.Scheme
}

//HasScheme >
func HasScheme(input string) bool {
	res := strings.Index(input, "://")
	return res >= 0
}

//RemoveProtocol removes protocol from target (something://...)
func RemoveProtocol(input string) string {
	res := strings.Index(input, "://")
	if res >= 0 {
		return input[res+3:]
	}
	return input
}

//RemovePort removes port from target (:80...)
func RemovePort(input string) string {
	res := strings.Index(input, ":")
	if res >= 0 {
		return input[:res-1]
	}
	return input
}

//RetrieveParameters from url
func RetrieveParameters(input string) []string {
	var result []string
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

//AbsoluteURL takes as input a path and returns the full
//absolute URL with protocol + host + path
func AbsoluteURL(protocol string, target string, path string) string {
	// if the path variable starts with a scheme, it means that the
	// path is itself an absolute path.
	if HasScheme(path) {
		return path
	}
	if len(path) != 0 && path[0] == '/' {
		return protocol + "://" + target + path
	}
	return protocol + "://" + target + "/" + path
}
