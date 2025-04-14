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

	@Author:      edoardottt, https://www.edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package output

import (
	"log"
	"os"
	"strings"

	fileUtils "github.com/edoardottt/cariddi/internal/file"
)

// AppendOutputToTxt opens the output file and append
// the string taken as input.
func AppendOutputToTxt(output string, filename string) {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, fileUtils.Permission0644)
	if err != nil {
		log.Println(err)
	}

	if _, err := file.WriteString(output + "\n"); err != nil {
		log.Fatal(err)
	}

	file.Close()
}

// WriteAllTxt opens the output file and append all the strings
// taken as input in one-shot operation.
func WriteAllTxt(output []string, filename string) {
	var buf strings.Builder

	for _, elem := range output {
		buf.WriteString(elem + "\n")
	}

	if err := os.WriteFile(filename, []byte(buf.String()), fileUtils.Permission0644); err != nil {
		log.Fatal(err)
	}
}
