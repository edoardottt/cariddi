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
	extension string
	severity  int
	alert     bool
}

func GetExtensions() []FileType {
	//extensions contains a list of known extensions
	//and the TYPICAL (also say `in general`) associated severity.
	var extensions = []FileType{
		{"key", 1, true},
		{"env", 1, true},
		{"pem", 1, true},
		{"git", 1, true},
		{"ovpn", 1, true},
		{"log", 1, true},
		{"secret", 1, true},
		{"secrets", 1, true},
		{"access", 1, true},
		{"bak", 1, true},
		{"dat", 1, true},
		{"db", 1, true},
		{"sh", 2, true},
		{"py", 2, true},
		{"json", 2, true},
		{"xml", 2, true},
		{"yml", 2, true},
		{"yaml", 2, true},
		{"properties", 2, true},
		{"toml", 2, true},
		{"dtd", 2, true},
		{"php4", 2, true},
		{"conf", 2, true},
		{"config", 2, true},
		{"configs", 2, true},
		{"zip", 3, true},
		{"doc", 3, false},
		{"docx", 3, false},
		{"dochtml", 3, false},
		{"csv", 3, false},
		{"odt", 3, false},
		{"xls", 3, false},
		{"xlsx", 3, false},
		{"txt", 3, false},
		{"ts", 4, false},
		{"js", 4, false},
		{"php", 5, false},
		{"asp", 5, false},
		{"jsp", 5, false},
		{"phtml", 5, false},
		{"php5", 5, false},
		{"html", 6, false},
		{"htm", 6, false},
		{"pdf", 7, false},
	}

	return extensions
}
