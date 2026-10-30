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

package scanner_test

import (
	"math/rand"
	"os"
	"regexp"
	"regexp/syntax"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"

	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

// The tests check that the engine returns exactly what running every regex
// on the whole text returns (the behavior before the engine existed).
// Set CARIDDI_EQUIVALENCE_ITERATIONS to run the randomized tests longer.

const (
	defaultIterations = 300
	raceDivisor       = 5 // fewer cases with -race (see raceEnabled).
	modeWindows       = "windows"
	modeFull          = "full"
)

// iterations returns how many cases the randomized tests check.
func iterations(t *testing.T) int {
	t.Helper()

	if v := os.Getenv("CARIDDI_EQUIVALENCE_ITERATIONS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("invalid CARIDDI_EQUIVALENCE_ITERATIONS: %v", err)
		}

		return n
	}

	if raceEnabled {
		return defaultIterations / raceDivisor
	}

	return defaultIterations
}

// fullScan is the reference: every regex run on the whole text.
func fullScan(regexes []*regexp.Regexp, text string) [][]string {
	result := make([][]string, len(regexes))

	for i, re := range regexes {
		for _, match := range re.FindAllStringSubmatch(text, -1) {
			result[i] = append(result[i], match[0])
		}
	}

	return result
}

// checkEngine fails if e and the full scan of regexes find different matches in text.
func checkEngine(t *testing.T, e *scanner.Engine, regexes []*regexp.Regexp, text string) {
	t.Helper()

	want := fullScan(regexes, text)
	got := e.FindAll(text)

	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d", len(got), len(want))
	}

	for i := range want {
		if !slices.Equal(want[i], got[i]) {
			t.Fatalf("regex %q on text %q:\nwant %q\ngot  %q", regexes[i], text, want[i], got[i])
		}
	}
}

// mustCompile compiles the patterns.
func mustCompile(t *testing.T, patterns ...string) []*regexp.Regexp {
	t.Helper()

	regexes := make([]*regexp.Regexp, 0, len(patterns))

	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			t.Fatalf("compile %q: %v", pattern, err)
		}

		regexes = append(regexes, re)
	}

	return regexes
}

// mustParse returns the simplified syntax tree of pattern.
func mustParse(t *testing.T, pattern string) *syntax.Regexp {
	t.Helper()

	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		t.Fatal(err)
	}

	return re.Simplify()
}

// generate appends to sb a random string matched (most of the time) by re.
func generate(rng *rand.Rand, re *syntax.Regexp, sb *strings.Builder) {
	switch re.Op { //nolint: exhaustive // other operators generate nothing.
	case syntax.OpLiteral:
		for _, r := range re.Rune {
			if re.Flags&syntax.FoldCase != 0 { // any rune of the case folding orbit.
				for range rng.Intn(3) {
					r = unicode.SimpleFold(r)
				}
			}

			sb.WriteRune(r)
		}
	case syntax.OpCharClass:
		i := 2 * rng.Intn(len(re.Rune)/2)
		lo, hi := re.Rune[i], re.Rune[i+1]

		if hi-lo > 200 {
			hi = lo + 200
		}

		sb.WriteRune(lo + rune(rng.Intn(int(hi-lo)+1)))
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		sb.WriteByte("abcXYZ019_-.= \"'"[rng.Intn(16)])
	case syntax.OpCapture:
		generate(rng, re.Sub[0], sb)
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		lo, hi := 0, 1

		switch re.Op { //nolint: exhaustive // only repetition operators get here.
		case syntax.OpStar:
			hi = 4
		case syntax.OpPlus:
			lo, hi = 1, 6
		case syntax.OpRepeat:
			lo, hi = re.Min, re.Max
			if hi < 0 || hi > lo+6 {
				hi = lo + 6
			}
		}

		for range lo + rng.Intn(hi-lo+1) {
			generate(rng, re.Sub[0], sb)
		}
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			generate(rng, sub, sb)
		}
	case syntax.OpAlternate:
		generate(rng, re.Sub[rng.Intn(len(re.Sub))], sb)
	}
}

// noise used around planted strings: delimiters, keywords alone (decoys),
// mixed case, invalid UTF-8, multi-byte runes, the Kelvin sign and the
// long s (that (?i) matches for k and s).
var noise = []string{ //nolint: gochecknoglobals
	"\"", "'", " ", "\n", "\r\n", "=", ":", "`", ";", ",", "(", ")", "<", ">", "/", ".", "-", "_",
	"a", "Z", "9", "é", "€", "𐐨", "\xff", "\xe2\x84", "\xc3", "\t", "\x00", "",
	"pat", "Pat", "xpat", "fb", "FB", "bs", "aws", "AWS", "key-", "-us", "akia", "AKIA",
	"twitter", "linkedin", "discord", "stripe", "heroku", "jira", "_aws", "9fb", "amazonaws.com",
	"s3://", "ghs_", "xoxb-", "sk-ant-api", "eyJ1Ijoi", "pk.", "-----BEGIN ", "glpat-",
	"warning", "Warning: ", "driver", "error", "Error", "exception", "SQL", "oracle", "ORA-", "mysql",
	"PostgreSQL", "stack trace", "Traceback", "\\W", "&lt;", "[SQL Server]", "db2_", "fatal",
	"\u212A", "\u017F", "A\u212AIA", "\u017Fk-ant-api", "a\u017Fana", "\u212Aey-",
}

// randomNoise appends n random noise strings to sb.
func randomNoise(rng *rand.Rand, sb *strings.Builder, n int) {
	for range n {
		sb.WriteString(noise[rng.Intn(len(noise))])
	}
}

// randomBody plants strings generated from trees among random noise.
func randomBody(rng *rand.Rand, trees []*syntax.Regexp) string {
	var sb strings.Builder

	if rng.Intn(4) != 0 { // sometimes start the body with a planted string.
		randomNoise(rng, &sb, rng.Intn(40))
	}

	for range 1 + rng.Intn(6) {
		generate(rng, trees[rng.Intn(len(trees))], &sb)
		randomNoise(rng, &sb, rng.Intn(8))

		if rng.Intn(20) == 0 {
			sb.WriteString(strings.Repeat("x", rng.Intn(2000)))
		}
	}

	if rng.Intn(20) == 0 {
		sb.WriteRune([]rune{'\u212A', '\u017F'}[rng.Intn(2)])
	}

	if rng.Intn(3) == 0 { // sometimes end the body with a planted string.
		randomNoise(rng, &sb, rng.Intn(40))
	}

	return sb.String()
}

// randomRegex returns a random (possibly invalid) regular expression.
func randomRegex(rng *rand.Rand, depth int) string {
	literals := []string{
		"ab", "key", "aws", "Kk", "SK", "sk", "pat", "fb", "xox", "ghp_", "é", "\u017F", "K", "𐐀", "ß",
		"-", `\.`, "_", "9", "=", `"`, "'", "@", "warning",
	}
	classes := []string{
		"[a-z]", "[A-F0-9]", "[0-9a-zA-Z_-]", `[^"' ]`, `\w`, `\d`, `\s`, `\S`, ".", `[\x{80}-\x{10FFFF}]`,
		"[ks]", "[a-z0-9.-]", `[^\n]`, `\pL`, "[[:alpha:]]", `[^&lt;&quot;]`,
	}
	assertions := []string{`\b`, `\B`, "^", "$", `\A`, `\z`}
	quantifiers := []string{"?", "*", "+", "{2}", "{1,3}", "{0,2}", "{2,}", "??", "*?", "+?"}
	flags := []string{"(?i:", "(?s:", "(?U:", "(?-i:", "(?m:", "(?is:"}

	if depth <= 0 || rng.Intn(4) == 0 {
		switch n := rng.Intn(10); {
		case n < 5:
			return literals[rng.Intn(len(literals))]
		case n < 9:
			return classes[rng.Intn(len(classes))]
		default:
			return assertions[rng.Intn(len(assertions))]
		}
	}

	switch rng.Intn(6) {
	case 0:
		return randomRegex(rng, depth-1) + "|" + randomRegex(rng, depth-1)
	case 1:
		return "(" + randomRegex(rng, depth-1) + ")"
	case 2:
		return "(?:" + randomRegex(rng, depth-1) + ")" + quantifiers[rng.Intn(len(quantifiers))]
	case 3:
		return flags[rng.Intn(len(flags))] + randomRegex(rng, depth-1) + ")"
	default:
		var sb strings.Builder
		for range 2 + rng.Intn(3) {
			sb.WriteString(randomRegex(rng, depth-1))
		}

		return sb.String()
	}
}

// countWindows returns how many regexes of e are run on windows of text.
func countWindows(e *scanner.Engine, text string) int {
	n := 0

	for _, mode := range scanner.EngineModes(e, text) {
		if mode == modeWindows {
			n++
		}
	}

	return n
}
