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

package output

import (
	"log"
	"os"

	fileUtils "github.com/edoardottt/cariddi/internal/file"
)

// WriteAllTxt opens the output file and append all the strings
// taken as input in one-shot operation.
func WriteAllTxt(output []string, filename string) {
	// One allocation of the exact size and one write.
	size := 0
	for _, elem := range output {
		size += len(elem) + 1
	}

	buf := make([]byte, 0, size)
	for _, elem := range output {
		buf = append(buf, elem...)
		buf = append(buf, '\n')
	}

	if err := os.WriteFile(filename, buf, fileUtils.Permission0644); err != nil {
		log.Fatal(err)
	}
}
