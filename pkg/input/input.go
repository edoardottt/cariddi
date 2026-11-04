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
	"maps"
	"os"
	"slices"
	"strings"

	fileUtils "github.com/edoardottt/cariddi/v2/internal/file"
	sliceUtils "github.com/edoardottt/cariddi/v2/internal/slice"
	pdutils "github.com/projectdiscovery/utils/file"
	"golang.org/x/net/http/httpguts"
)

const (
	// maxTargetLength is the maximum length of an input line, with its
	// line break: a longer line can't be a real URL.
	maxTargetLength = 64 * 1024
	// skippedPreviewLength is how much of a skipped line is printed.
	skippedPreviewLength = 100
)

// ScanTargets returns the targets: the URLs given with -u, then the lines
// of the list files (-l), in order, then the lines of stdin (when it isn't
// a terminal), without duplicates. Without any target, it exits.
// The lines longer than maxTargetLength (with the line break) are skipped
// (and printed in debug mode), the following ones are still read.
func ScanTargets(urls, lists []string, debug bool) []string {
	stdin := pdutils.HasStdin()

	if len(urls) == 0 && len(lists) == 0 && !stdin {
		fmt.Println("No input provided: use -u, -l or pass the URLs on stdin.")
		os.Exit(1)
	}

	var result []string

	for _, url := range urls {
		if target := strings.TrimSpace(url); target != "" {
			result = append(result, target)
		}
	}

	for _, list := range lists {
		// Nothing to add (e.g. -l "$EMPTY").
		if list == "" {
			continue
		}

		file, err := os.Open(list)
		if err != nil {
			fmt.Println("Can't open the list of URLs:", err)
			os.Exit(1)
		}

		result = append(result, readTargets(file, list, debug)...)

		file.Close()
	}

	if stdin {
		result = append(result, readTargets(os.Stdin, "stdin", debug)...)
	}

	result = sliceUtils.RemoveDuplicateValues(result)

	// E.g. -u "" or an empty file: there's nothing to crawl.
	if len(result) == 0 {
		fmt.Println("No URLs to crawl.")
		os.Exit(1)
	}

	return result
}

// readTargets returns the targets read from r (named name in the
// messages), one per line.
func readTargets(r io.Reader, name string, debug bool) []string {
	var result []string

	// The targets are not lowercased: the path and
	// the query of a URL are case sensitive.
	reader := bufio.NewReaderSize(r, maxTargetLength)

	for lineNumber := 1; ; lineNumber++ {
		line, err := reader.ReadSlice('\n')

		if errors.Is(err, bufio.ErrBufferFull) {
			// The line is too long: skip it up to its end.
			skipped := len(line)
			preview := string(line[:skippedPreviewLength])

			for errors.Is(err, bufio.ErrBufferFull) {
				line, err = reader.ReadSlice('\n')
				skipped += len(line)
			}

			// The line break is not part of the line.
			skipped -= len(line) - len(bytes.TrimRight(line, "\r\n"))

			// The preview is quoted: the line can contain control characters.
			if debug {
				log.Printf("Line %d of %s skipped, too long for a URL (%d bytes): %q...",
					lineNumber, name, skipped, preview)
			}

			line = nil
		}

		if target := strings.TrimSpace(string(line)); target != "" {
			result = append(result, target)
		}

		if err != nil {
			// A read error must not leave the input silently incomplete.
			if !errors.Is(err, io.EOF) {
				fmt.Printf("Error reading %s: %v\n", name, err)
				os.Exit(1)
			}

			break
		}
	}

	return result
}

// isFile reports whether value is the name of a file to read: anything
// but a directory (e.g. also a named pipe, as <(cat headers.txt)).
func isFile(value string) bool {
	info, err := os.Stat(value)

	return err == nil && !info.IsDir()
}

// GetHeaders returns the headers provided as input using the headers
// flag. Each value is a file containing headers, or the headers
// themselves: in both, ";;" and the line breaks separate them,
// e.g. -H "Cookie: auth=yes;;Client: type=2" or one per line in a file.
// A header found more than once keeps its last value.
// A value without headers, or with an invalid header name, makes it exit.
func GetHeaders(values []string) map[string]string {
	result := make(map[string]string)

	for _, value := range values {
		// Nothing to add (e.g. -H "$EMPTY").
		if strings.TrimSpace(value) == "" {
			continue
		}

		content, source := value, fmt.Sprintf("The headers value %q is not a file and", value)
		if isFile(value) {
			content, source = string(fileUtils.ReadEntireFile(value)), "The headers file "+value
		}

		headers := parseHeaders(content)
		if len(headers) == 0 {
			fmt.Println(source + " doesn't contain any header (Name: value).")
			os.Exit(1)
		}

		// The HTTP client would refuse every request with an invalid name
		// (e.g. with a space). Sorted, the same name is always reported.
		for _, name := range slices.Sorted(maps.Keys(headers)) {
			if !httpguts.ValidHeaderFieldName(name) {
				fmt.Printf("%s contains an invalid header name: %q.\n", source, name)
				os.Exit(1)
			}
		}

		maps.Copy(result, headers)
	}

	return result
}

// parseHeaders returns the headers contained in input, separated by ";;"
// or by line breaks: a header can't contain them.
func parseHeaders(input string) map[string]string {
	result := make(map[string]string)

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

	return result
}

// GetIgnore returns the strings provided as input using the ignore flag:
// the URLs containing one of them are not crawled. Each value is a file
// (one string per line) or a comma-separated list. The strings are
// trimmed and the blank ones are skipped: they would match every URL.
func GetIgnore(values []string) []string {
	var result []string

	for _, value := range values {
		if isFile(value) {
			// ReadFile skips the blank lines.
			for _, line := range fileUtils.ReadFile(value) {
				result = append(result, strings.TrimSpace(line))
			}
		} else {
			result = append(result, sliceUtils.CheckInputArray(value)...)
		}
	}

	return sliceUtils.RemoveDuplicateValues(result)
}

// GetProxies returns the proxies provided as input using the proxy flag.
// Each value is a proxy or a file containing proxies, one per line (trimmed,
// the blank lines are skipped). The order is kept: the requests use the
// proxies in turn, so a proxy given twice gets twice the requests. An empty
// value adds nothing (e.g. -proxy "$EMPTY").
// An invalid proxy or a file without proxies makes it exit: the requests
// would fail, or not use the proxies.
// Without proxies, it checks the proxy environment variables, used instead.
func GetProxies(values []string) []string {
	var result []string

	for _, value := range values {
		if value == "" {
			continue
		}

		if !isFile(value) {
			if !CheckProxy(value) {
				fmt.Printf("The proxy %q is not a file nor a URL with scheme http, https, socks5 or socks5h.\n", value)
				fmt.Println("Examples:")
				fmt.Println("	- cat urls | cariddi -proxy http://127.0.0.1:8080")
				fmt.Println("	- cat urls | cariddi -proxy socks5://127.0.0.1:1080")
				fmt.Println("	- cat urls | cariddi -proxy proxies.txt")
				os.Exit(1)
			}

			result = append(result, value)

			continue
		}

		// ReadFile skips the blank lines.
		lines := fileUtils.ReadFile(value)
		if len(lines) == 0 {
			fmt.Printf("The proxies file %s doesn't contain any proxy (one per line).\n", value)
			os.Exit(1)
		}

		for _, line := range lines {
			proxy := strings.TrimSpace(line)
			if !CheckProxy(proxy) {
				fmt.Printf("The proxies file %s contains an invalid proxy: %q.\n", value, proxy)
				fmt.Println("Each line must be a URL with scheme http, https, socks5 or socks5h (e.g. socks5://127.0.0.1:1080).")
				os.Exit(1)
			}

			result = append(result, proxy)
		}
	}

	if len(result) == 0 {
		checkProxyEnvironment()
	}

	return result
}
