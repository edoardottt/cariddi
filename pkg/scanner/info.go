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

package scanner

import (
	"regexp"
	"sync"
)

// Info struct.
// Name = the name that identifies the information.
// Regex = The regular expression to be matched.
// FalsePositive = if not nil, it reports whether the match
// text[start:end] is a false positive, given its context.
type Info struct {
	Name          string
	Regex         regexp.Regexp
	FalsePositive func(text string, start, end int) bool
}

// InfoMatched struct.
// Info = Info struct.
// Url = url in which the information is found.
// Match = the string matching the regex.
type InfoMatched struct {
	Info  Info
	URL   string
	Match string
}

var (
	infos     []Info    //nolint: gochecknoglobals
	onceInfos sync.Once //nolint: gochecknoglobals
	// escapeSuffix matches an escape sequence of a JSON or JavaScript
	// string at the end of a text, e.g. \n, \t, \u003e or \x3e.
	escapeSuffix = regexp.MustCompile(`\\(?:[nrt]|u[0-9a-fA-F]{4}|x[0-9a-fA-F]{2})$`)
)

// GetInfoRegexes returns all the info structs.
func GetInfoRegexes() []Info {
	onceInfos.Do(func() {
		infos = []Info{
			{
				"Email address",
				*regexp.MustCompile(`(?i)([a-zA-Z0-9_.+-]+@[a-zA-Z0-9]+[a-zA-Z0-9-]*\.[a-zA-Z0-9-.]*[a-zA-Z0-9]{2,})`),
				nil,
			},
			{
				// As browsers parse them: any content, closed by the first
				// "-->" or "--!>", and the empty "<!-->" and "<!--->".
				"HTML comment",
				*regexp.MustCompile(`(?s)<!--(?:-?>|.*?--!?>)`),
				nil,
			},
			{
				"IPv4 address",
				*regexp.MustCompile(`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}\b`),
				ipv4FalsePositive,
			},
			/*
				TOO MANY FALSE POSITIVES
				{
					"BTC address",
					`([13]|bc1)[A-HJ-NP-Za-km-z1-9]{27,34}`,
				},
			*/
			/*
				HOW TO AVOID VERY VERY LONG BASE64 IMAGES ???
				{
					"Base64-encoded JSON",
					`ey(A|B)[A-Za-z0-9+\/]{20,}(={0,2})`,
				},
			*/
		}
	})

	return infos
}

// FindAll returns the matches of the info regex in text,
// excluding the false positives.
func (i *Info) FindAll(text string) []string {
	var matches []string

	for _, loc := range i.Regex.FindAllStringIndex(text, -1) {
		if i.FalsePositive == nil || !i.FalsePositive(text, loc[0], loc[1]) {
			matches = append(matches, text[loc[0]:loc[1]])
		}
	}

	return matches
}

// ipv4FalsePositive reports whether text[start:end] is part of a longer
// word or number, e.g. 1.2.3.4 in v1.2.3.4 or 11.2.3.4, or of a longer
// dotted number, e.g. in the version 1.2.3.4.5 or the OID 1.3.6.1.4.1
// (the dot of a sentence end is not part of it). An escape sequence is not
// part of a word, e.g. \n10.0.0.1 or \u003e10.0.0.1 in a JSON string.
func ipv4FalsePositive(text string, start, end int) bool {
	// at returns the byte at i, 0 if i is out of text.
	at := func(i int) byte {
		if i < 0 || i >= len(text) {
			return 0
		}

		return text[i]
	}
	isDigit := func(c byte) bool { return '0' <= c && c <= '9' }
	isWord := func(c byte) bool { return isDigit(c) || c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' }
	before, after := start-1, end

	return (isWord(at(before)) && !escapeSuffix.MatchString(text[max(0, start-len(`\u0000`)):start])) ||
		(at(before) == '.' && isDigit(at(before-1))) ||
		(at(after) == '.' && isDigit(at(after+1)))
}

// RemoveDuplicateInfos removes duplicates from Infos found.
func RemoveDuplicateInfos(input []InfoMatched) []InfoMatched {
	keys := make(map[string]bool)
	list := []InfoMatched{}

	for _, entry := range input {
		if _, value := keys[entry.Match]; !value {
			keys[entry.Match] = true
			list = append(list, entry)
		}
	}

	return list
}
