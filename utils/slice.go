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
	"strings"
)

//RemoveDuplicateValues removes duplicates rom a string slice
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
