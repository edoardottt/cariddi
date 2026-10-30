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

package slice

import (
	"strings"
)

// RemoveDuplicateValues removes duplicates from a slice
// of strings.
func RemoveDuplicateValues(strSlice []string) []string {
	keys := make(map[string]bool)
	list := make([]string, 0, len(strSlice))

	for _, entry := range strSlice {
		if ok := keys[entry]; !ok {
			keys[entry] = true
			list = append(list, entry)
		}
	}

	return list
}

// CheckInputArray checks the basic rules to
// be valid and then returns the array as input.
// - Trim the spaces around the elements (e.g. "forum, blog").
// - Delete duplicates.
// - Avoid empty strings.
func CheckInputArray(input string) []string {
	delimiter := byte(',')
	sliceOut := strings.Split(input, string(delimiter))
	result := make([]string, 0, len(sliceOut))

	for _, elem := range sliceOut {
		if elem = strings.TrimSpace(elem); elem != "" {
			result = append(result, elem)
		}
	}

	return RemoveDuplicateValues(result)
}

// Contains checks if a string str is contained in the input slice.
func Contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}

	return false
}
