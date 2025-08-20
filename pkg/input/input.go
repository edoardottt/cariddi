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

package input

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
	pdutils "github.com/projectdiscovery/utils/file"
)

const (
	coupleSize = 2
)

// ScanTargets return the array of elements
// taken as input on stdin.
func ScanTargets() []string {
	if !pdutils.HasStdin() {
		fmt.Println("No input provided.")
		os.Exit(1)
	}

	var result []string

	// accept domains on stdin
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domain := strings.ToLower(sc.Text())
		if len(domain) > coupleSize {
			result = append(result, domain)
		}
	}

	return sliceUtils.RemoveDuplicateValues(result)
}

// GetHeaders returns the headers provided as input
// using the headers flag.
// E.g. -headers \"Cookie: auth=yes;;Client: type=2\".
func GetHeaders(input string) map[string]string {
	result := make(map[string]string)

	if input != "" {
		if !strings.Contains(input, ":") {
			fmt.Println("The headers provided don't contains the : separator.")
			os.Exit(1)
		}

		headers := strings.Split(input, ";;")
		for _, header := range headers {
			var parts []string
			if strings.Contains(header, ":") {
				parts = strings.SplitN(header, ":", coupleSize)
			} else {
				continue
			}

			result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
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
