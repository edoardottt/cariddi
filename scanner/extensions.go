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

//FileType
type FileType struct {
	Extension string
	Severity  int
}

//FileTypeMatched
type FileTypeMatched struct {
	Filetype FileType
	Url      string
}

func GetExtensions() []FileType {
	//extensions contains a list of known extensions
	//and the TYPICAL (also say `in general`) associated severity.
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
		{"properties", 2},
		{"dtd", 2},
		{"conf", 2},
		{"config", 2},
		{"configs", 2},
		{"sh", 3},
		{"py", 3},
		{"txt", 3},
		{"xml", 3},
		{"yml", 3},
		{"yaml", 3},
		{"toml", 3},
		{"php4", 3},
		{"json", 3},
		{"zip", 3},
		{"doc", 3},
		{"docx", 3},
		{"dochtml", 3},
		{"csv", 3},
		{"odt", 3},
		{"xls", 3},
		{"xlsx", 3},
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
