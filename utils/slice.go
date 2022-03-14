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

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package utils

import (
	"net/http"
	"strings"
)

//RemoveDuplicateValues removes duplicates from a slice
//of strings
func RemoveDuplicateValues(strSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range strSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

//CheckInputArray checks the basic rules to
//be valid and then returns the array as input.
// - Delete duplicates
// - Avoid empty strings
func CheckInputArray(input string) []string {
	delimiter := byte(',')
	sliceOut := strings.Split(input, string(delimiter))
	sliceOut = RemoveDuplicateValues(sliceOut)
	result := []string{}
	for _, elem := range sliceOut {
		if elem != "" {
			result = append(result, elem)
		}
	}
	return result
}

//CheckCookies checks if the string provided to the
//-cookie option is valid.
//format: "name1:value1;name2:value2"
//It returns a slice of Cookies
func CheckCookies(input string) []*http.Cookie {
	var result []*http.Cookie
	if input == "" {
		return result
	}
	//Split and get different pairs of (name,value)
	pairs := strings.Split(input, ";")
	if len(pairs) == 0 {
		return result
	}
	for _, pair := range pairs {
		couple := strings.Split(pair, ":")
		if len(couple) == 0 {
			continue
		}
		if len(couple) == 2 {
			result = append(result, &http.Cookie{Name: couple[0], Value: couple[1]})
		} else {
			continue
		}
	}
	return result
}
