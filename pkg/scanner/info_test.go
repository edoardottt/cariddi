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
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"
	"testing"

	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

// infosFullScan is the original infos hunting (every regex on the whole
// body, false positives removed), run sequentially in rule order.
func infosFullScan(infos []scanner.Info, body string) []scanner.InfoMatched {
	var result []scanner.InfoMatched

	for _, info := range infos {
		for _, loc := range info.Regex.FindAllStringIndex(body, -1) {
			match := body[loc[0]:loc[1]]
			isFalsePositive := info.FalsePositive != nil && info.FalsePositive(body, loc[0], loc[1])

			for _, fp := range info.FalsePositives {
				isFalsePositive = isFalsePositive || strings.Contains(strings.ToLower(match), fp)
			}

			if !isFalsePositive {
				result = append(result, scanner.InfoMatched{Info: info, URL: testTarget, Match: match})
			}
		}
	}

	return result
}

// checkInfos fails if s and the original infos hunting find different infos in body.
func checkInfos(t *testing.T, s *scanner.InfoScanner, infos []scanner.Info, body string) {
	t.Helper()

	want := infosFullScan(infos, body)
	got := s.Scan(testTarget, body)

	same := len(want) == len(got)
	for i := 0; same && i < len(want); i++ {
		w, g := want[i], got[i]
		same = w.Match == g.Match && w.URL == g.URL && w.Info.Name == g.Info.Name &&
			w.Info.Regex.String() == g.Info.Regex.String()
	}

	if !same {
		t.Fatalf("body %q:\nwant %v\ngot  %v", body, infoNames(want), infoNames(got))
	}
}

// infoNames describes matches, for test failures.
func infoNames(matches []scanner.InfoMatched) []string {
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		result = append(result, m.Info.Name+": "+strconv.Quote(m.Match))
	}

	return result
}

// TestBuiltinInfosHaveKeywords checks that no predefined info regex runs on every body.
func TestBuiltinInfosHaveKeywords(t *testing.T) {
	keywords := scanner.EngineKeywords(scanner.InfoScannerEngine(scanner.GetInfoScanner()))

	for i, info := range scanner.GetInfoRegexes() {
		if len(keywords[i]) == 0 {
			t.Errorf("no keywords derived for %q: it will be run on every body, add a literal to the regex", info.Name)
		}
	}
}

// TestInfoScannerEdgeCases checks the predefined infos on tricky bodies,
// above all the IPv4 false positives, which depend on the bytes around
// the match (also when the regex runs on windows of the body).
func TestInfoScannerEdgeCases(t *testing.T) {
	infos := scanner.GetInfoRegexes()
	s := scanner.GetInfoScanner()
	filler := strings.Repeat("lorem ipsum ", 200)

	// URL userinfo (also with escaped slashes), an escape, an asset name, SSH
	// locations and an address before a port.
	emailContexts := `https://u:p@example.com/ http:\/\/k@sentry.io \u003cx@example.com logo@2x.png ` +
		`git@github.com:org/repo.git ssh://git@example.com/r a@example.com:587 `

	bodies := map[string]string{
		"empty":                 "",
		"ip at start and end":   "10.0.0.1 " + filler + " 192.168.1.254",
		"ip in a word":          filler + "v1.2.3.4 x10.0.0.1 10.0.0.1x " + filler,
		"ip in a dotted number": filler + "1.2.3.4.5 OID 1.3.6.1.4.1 end 10.0.0.1. " + filler,
		"ip after escapes":      filler + `\n10.0.0.1 \t10.0.0.2 \u003e10.0.0.3 \x3e10.0.0.4 a\10.0.0.5` + filler,
		"ip near multibyte":     "€10.0.0.1€ é192.168.0.1é",
		"dense dots (cap)":      strings.Repeat("a.", 5000) + " 10.0.0.1 " + strings.Repeat("1.", 3000),
		"emails":                filler + "mail a@b.com, x@@y.com, A.B+c@Sub.Example.ORG " + filler,
		"dense at signs (cap)":  strings.Repeat("@", 3000) + "a@b.com",
		"email contexts":        filler + emailContexts + filler + "b@example.com",
		"html comments":         "<!--> <!---> <!-- a -- b --> <!-- x --!> <!-- unclosed",
		"comment with ip":       "<!-- 10.0.0.1 admin@example.com -->",
		"invalid utf8":          "\xff10.0.0.1\xfe a@b.com\xe2\x84",
		"kelvin sign":           "\u212A@b.com k@\u212A.com",
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			checkInfos(t, s, infos, body)
		})
	}
}

// TestInfoScannerCustomInfos checks both kinds of false positives,
// together and alone, and an empty list.
func TestInfoScannerCustomInfos(t *testing.T) {
	notAfterX := func(text string, start, _ int) bool { return start > 0 && text[start-1] == 'x' }
	infos := []scanner.Info{
		{Name: "a", Regex: *regexp.MustCompile(`key-[a-z]{4}`), FalsePositives: []string{"test"}},
		{Name: "b", Regex: *regexp.MustCompile(`(?i)KEY-[a-z]{4}`), FalsePositive: notAfterX},
		{Name: "c", Regex: *regexp.MustCompile(`key-[a-z]{4}`), FalsePositives: []string{"abcd"}, FalsePositive: notAfterX},
	}
	s := scanner.NewInfoScanner(infos)
	body := "key-abcd key-TEST xkey-wxyz KEY-efgh"

	checkInfos(t, s, infos, body)

	if got := infoNames(s.Scan(testTarget, body)); strings.Join(got, ",") !=
		`a: "key-abcd",a: "key-wxyz",b: "key-abcd",b: "key-TEST",b: "KEY-efgh"` {
		t.Fatalf("got %v", got)
	}

	if got := scanner.NewInfoScanner(nil).Scan(testTarget, body); len(got) != 0 {
		t.Fatalf("empty scanner found %v", infoNames(got))
	}
}

// infoTrees returns the syntax trees of the predefined info regexes.
func infoTrees(t *testing.T) []*syntax.Regexp {
	t.Helper()

	infos := scanner.GetInfoRegexes()
	trees := make([]*syntax.Regexp, 0, len(infos))

	for _, info := range infos {
		trees = append(trees, mustParse(t, info.Regex.String()))
	}

	return trees
}

// TestInfoScannerRandom plants random strings generated from the
// predefined regexes in random noise.
func TestInfoScannerRandom(t *testing.T) {
	infos := scanner.GetInfoRegexes()
	s := scanner.GetInfoScanner()
	trees := infoTrees(t)
	rng := rand.New(rand.NewSource(7))
	windows, matches := 0, 0

	for range iterations(t) {
		body := randomBody(rng, trees)
		checkInfos(t, s, infos, body)

		windows += countWindows(scanner.InfoScannerEngine(s), body)
		matches += len(s.Scan(testTarget, body))
	}

	// Make sure the interesting code paths are exercised.
	if windows < iterations(t) || matches < iterations(t) {
		t.Fatalf("weak test: %d windowed scans, %d matches", windows, matches)
	}
}

// FuzzInfoScanner checks the predefined infos on random bodies.
func FuzzInfoScanner(f *testing.F) {
	f.Add("10.0.0.1 v1.2.3.4 1.2.3.4.5 \\n10.0.0.2")
	f.Add("a@b.com <!-- x --> <!-->")
	f.Add("\u212A@b.com")

	infos := scanner.GetInfoRegexes()
	s := scanner.GetInfoScanner()

	f.Fuzz(func(t *testing.T, body string) {
		checkInfos(t, s, infos, body)
	})
}

// BenchmarkInfoScanner measures the infos hunting.
func BenchmarkInfoScanner(b *testing.B) {
	body := benchmarkBody()
	s := scanner.GetInfoScanner()

	b.SetBytes(int64(len(body)))

	for b.Loop() {
		s.Scan(testTarget, body)
	}
}

// BenchmarkInfosFullScan measures the original infos hunting, for comparison.
func BenchmarkInfosFullScan(b *testing.B) {
	body := benchmarkBody()
	infos := scanner.GetInfoRegexes()

	b.SetBytes(int64(len(body)))

	for b.Loop() {
		infosFullScan(infos, body)
	}
}
