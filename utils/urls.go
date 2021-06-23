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
)

//GetHost >
func GetHost(input string) string {
	u, err := url.Parse(input)
	if err != nil {
		return ""
	}
	return u.Host
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
