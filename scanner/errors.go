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

//Error struct
// ErrorName = the name that identifies the error
// Regex = The regular expression to be matched
type Error struct {
	ErrorName string
	Regex     []string
}

//ErrorMatched struct
// Error = Error struct
// Url = url in which the error is found
// Match = the string matching the regex
type ErrorMatched struct {
	Error Error
	Url   string
	Match string
}

//GetErrorRegexes returns all the error structs
func GetErrorRegexes() []Error {
	var regexes = []Error{
		{
			"PHP error",
			[]string{`(?i)php warning`,
				`(?i)php error`,
				`(?i)include_path`,
				`(?i)undefined index`,
				`(?i)undefined variable`,
				`(?i)expect(s*) parameter [A-Za-z0-9-_]{1,30}`},
		},
		{
			"General error",
			[]string{`(?i)fatal error`,
				`(?i)critical error`,
				`(?i)severe error`,
				`(?i)high error`,
				`(?i)medium error`,
				`(?i)uncaught exception`},
		},
		{
			"Debug information",
			[]string{`(?i)Debug trace`, `(?i)stack trace\\:`},
		},
		{
			"MySQL error",
			[]string{`(?i)valid MySQL result`,
				`(?i)check the manual that (fits|corresponds to) your MySQL server version`,
				"(?i)MySQLSyntaxErrorException",
				"(?i)MySqlException",
				"(?i)MySql error",
				"(?i)Unknown column "},
		},
		{
			"MariaDB error",
			[]string{`(?i)check the manual that (fits|corresponds to) your MariaDB server version`,
				"(?i)MariaDB error"},
		},
		{
			"PostgreSQL error",
			[]string{`(?i)valid PostgreSQL result`,
				"(?i)PG::SyntaxError:",
				"(?i)PSQLException",
				"(?i)PostgreSQL query failed",
				"(?i)ERROR: parser: parse error at or near",
				"(?i)PostgreSQL error"},
		},
		{
			"MSSQL error",
			[]string{`(?i)Microsoft SQL error`,
				"(?i)Microsoft SQL Native Client error",
				"(?i)ODBC SQL Server Driver",
				"(?i)Unclosed quotation mark after the character string",
				"(?i)SQLServer JDBC Driver"},
		},
	}
	return regexes
}

//RemoveDuplicateErrors removes duplicates from Errors found
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
