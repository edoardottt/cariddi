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
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
	pdutils "github.com/projectdiscovery/utils/file"
)

const (
	// maxTargetLength is the maximum length of an input line: a longer
	// line can't be a real URL.
	maxTargetLength = 64 * 1024
	// skippedPreviewLength is how much of a skipped line is printed.
	skippedPreviewLength = 100
)

// ScanTargets return the array of elements
// taken as input on stdin.
// The lines longer than maxTargetLength are skipped (and printed
// in debug mode), the following ones are still read.
func ScanTargets(debug bool) []string {
	if !pdutils.HasStdin() {
		fmt.Println("No input provided.")
		os.Exit(1)
	}

	var result []string

	// Accept the targets on stdin, one per line. They are not lowercased:
	// the path and the query of a URL are case sensitive.
	r := bufio.NewReaderSize(os.Stdin, maxTargetLength)

	for lineNumber := 1; ; lineNumber++ {
		line, err := r.ReadSlice('\n')

		if errors.Is(err, bufio.ErrBufferFull) {
			// The line is too long: skip it up to its end.
			skipped := len(line)
			preview := string(line[:skippedPreviewLength])

			for errors.Is(err, bufio.ErrBufferFull) {
				line, err = r.ReadSlice('\n')
				skipped += len(line)
			}

			// The line break is not part of the line.
			skipped -= len(line) - len(bytes.TrimRight(line, "\r\n"))

			// The preview is quoted: the line can contain control characters.
			if debug {
				log.Printf("Input line %d skipped, too long for a URL (%d bytes): %q...", lineNumber, skipped, preview)
			}

			line = nil
		}

		if target := strings.TrimSpace(string(line)); target != "" {
			result = append(result, target)
		}

		if err != nil {
			// A read error must not leave the input silently incomplete.
			if !errors.Is(err, io.EOF) {
				fmt.Println("Error reading the input:", err)
				os.Exit(1)
			}

			break
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
