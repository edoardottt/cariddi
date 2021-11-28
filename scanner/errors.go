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

package scanner

//Error struct
type Error struct {
	ErrorName string
	Regex     string
}

//ErrorMatched struct
type ErrorMatched struct {
	Error Error
	Url   string
	Match string
}

//GetErrorRegexes returns all the error regexes
func GetErrorRegexes() []Error {
	var regexes = []Error{
		{
			"PHP error",
			`(php warning|php error|include_path|undefined index|undefined variable|\\?php|expects parameter [0-9]*)`,
		},
		{
			"General error",
			`(((fatal|critical|severe|high|medium) error)|uncaught exception)`,
		},
		{
			"Debug information",
			`(Debug trace|stack trace\\:)`,
		},
	}
	return regexes
}

//RemoveDuplicateErrors removes duplicates from secrets found
func RemoveDuplicateErrors(input []ErrorMatched) []ErrorMatched {
	keys := make(map[string]bool)
	list := []ErrorMatched{}
	for _, entry := range input {
		if _, value := keys[entry.Match+entry.Url]; !value {
			keys[entry.Url] = true
			list = append(list, entry)
		}
	}
	return list
}
