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

package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
	pdutils "github.com/projectdiscovery/utils/file"
)

// ScanTargets return the array of elements
// taken as input on stdin.
func ScanTargets() []string {
	if !pdutils.HasStdin() {
		fmt.Println("No input provided.")
		os.Exit(1)
	}

	var result []string

	// Accept the targets on stdin, one per line. They are not lowercased:
	// the path and the query of a URL are case sensitive.
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		target := strings.TrimSpace(sc.Text())
		if target != "" {
			result = append(result, target)
		}
	}

	return sliceUtils.RemoveDuplicateValues(result)
}

// GetHeaders returns the headers provided as input
// using the headers flag.
// E.g. -headers \"Cookie: auth=yes;;Client: type=2\".
// Line breaks separate the headers too (e.g. one header
// per line in a headers file): a header can't contain them.
func GetHeaders(input string) map[string]string {
	result := make(map[string]string)

	if input != "" {
		if !strings.Contains(input, ":") {
			fmt.Println("The headers provided don't contains the : separator.")
			os.Exit(1)
		}

		lines := strings.FieldsFuncSeq(input, func(r rune) bool { return r == '\n' || r == '\r' })
		for line := range lines {
			for header := range strings.SplitSeq(line, ";;") {
				name, value, found := strings.Cut(header, ":")
				name = strings.TrimSpace(name)

				// Skip the headers without a name, e.g. the HTTP/2 pseudo
				// headers (":authority: ...") copied from a browser.
				if !found || name == "" {
					continue
				}

				result[name] = strings.TrimSpace(value)
			}
		}
	} else {
		fmt.Println("Headers or HeadersFile flag provided, but the content is empty.")
		os.Exit(1)
	}

	if len(result) == 0 {
		fmt.Println("Headers or HeadersFile flag provided, but the content is empty.")
		os.Exit(1)
	}

	return result
}
