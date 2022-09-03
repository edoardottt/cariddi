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

package scanner

//FileType struct.
// Extension = the file extension (doc, txt ..etc..).
// Severity = the 'importance' of the file found. Higher is better.
type FileType struct {
	Extension string
	Severity  int
}

//FileTypeMatched struct.
// Filetype = Filetype struct.
// Url = url of the file found.
type FileTypeMatched struct {
	Filetype FileType
	URL      string
}

//GetExtensions returns all the extension structs.
func GetExtensions() []FileType {
	//extensions contains a list of known extensions
	//and the TYPICAL (also say `in general`) associated severity.
	//Why in general? Because a python file can be anything, it can
	//contain secret data or not.
	var extensions = []FileType{
		{"key", 1},
		{"env", 1},
		{"pem", 1},
		{"git", 1},
		{"ovpn", 1},
		{"log", 1},
		{"secret", 1},
		{"secrets", 1},
		{"access", 1},
		{"bak", 1},
		{"dat", 1},
		{"db", 1},
		{"sql", 1},
		{"pwd", 1},
		{"passwd", 1},
		{"gitignore", 1},
		{"properties", 2},
		{"dtd", 2},
		{"conf", 2},
		{"cfg", 2},
		{"config", 2},
		{"configs", 2},
		{"apk", 2},
		{"cgi", 3},
		{"sh", 3},
		{"py", 3},
		{"txt", 3},
		{"xml", 3},
		{"java", 3},
		{"rb", 3},
		{"rs", 3},
		{"go", 3},
		{"yml", 3},
		{"yaml", 3},
		{"toml", 3},
		{"php4", 3},
		{"json", 3},
		{"zip", 3},
		{"tar", 3},
		{"gz", 3},
		{"dochtml", 3},
		{"doc", 4},
		{"docx", 4},
		{"csv", 4},
		{"odt", 4},
		{"xls", 4},
		{"xlsx", 4},
		{"ts", 4},
		{"js", 4},
		{"php", 5},
		{"asp", 5},
		{"jsp", 5},
		{"phtml", 5},
		{"php5", 5},
		{"html", 6},
		{"htm", 6},
		{"pdf", 7},
	}

	return extensions
}

//RemoveDuplicateExtensions removes duplicates from Extensions found.
func RemoveDuplicateExtensions(input []FileTypeMatched) []FileTypeMatched {
	keys := make(map[string]bool)
	list := []FileTypeMatched{}

	for _, entry := range input {
		if _, value := keys[entry.URL]; !value {
			keys[entry.URL] = true
			list = append(list, entry)
		}
	}

	return list
}
