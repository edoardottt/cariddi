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
	"regexp/syntax"
	"strconv"
	"strings"
	"testing"

	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

// errorsFullScan is the original errors hunting (every regex on the whole
// body), run sequentially in rule order.
func errorsFullScan(errs []scanner.Error, body string) []scanner.ErrorMatched {
	var result []scanner.ErrorMatched

	for _, e := range errs {
		for _, match := range e.Regex.FindAllStringSubmatch(body, -1) {
			result = append(result, scanner.ErrorMatched{Error: e, URL: testTarget, Match: match[0]})
		}
	}

	return result
}

// checkErrors fails if s and the original errors hunting find different errors in body.
func checkErrors(t *testing.T, s *scanner.ErrorScanner, errs []scanner.Error, body string) {
	t.Helper()

	want := errorsFullScan(errs, body)
	got := s.Scan(testTarget, body)

	same := len(want) == len(got)
	for i := 0; same && i < len(want); i++ {
		w, g := want[i], got[i]
		same = w.Match == g.Match && w.URL == g.URL && w.Error.ErrorName == g.Error.ErrorName &&
			w.Error.Regex.String() == g.Error.Regex.String()
	}

	if !same {
		t.Fatalf("body %q:\nwant %v\ngot  %v", body, errorNames(want), errorNames(got))
	}
}

// errorNames describes matches, for test failures.
func errorNames(matches []scanner.ErrorMatched) []string {
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		result = append(result, m.Error.ErrorName+": "+strconv.Quote(m.Match))
	}

	return result
}

// errorTrees returns the syntax trees of the predefined error regexes.
func errorTrees(t *testing.T) []*syntax.Regexp {
	t.Helper()

	errs := scanner.GetErrorRegexes()
	trees := make([]*syntax.Regexp, 0, len(errs))

	for _, e := range errs {
		trees = append(trees, mustParse(t, e.Regex.String()))
	}

	return trees
}

// TestBuiltinErrorsHaveKeywords checks that no predefined error regex runs on every body.
func TestBuiltinErrorsHaveKeywords(t *testing.T) {
	keywords := scanner.EngineKeywords(scanner.ErrorScannerEngine(scanner.GetErrorScanner()))

	for i, e := range scanner.GetErrorRegexes() {
		if len(keywords[i]) == 0 {
			t.Errorf("no keywords derived for %q: it will be run on every body, add a literal to the regex", e.ErrorName)
		}
	}
}

// maxAttempts bounds the strings generated to get one matched by a regex.
const maxAttempts = 100

// TestErrorScannerContexts generates strings matched by each predefined
// error regex and checks them in different surroundings.
func TestErrorScannerContexts(t *testing.T) {
	errs := scanner.GetErrorRegexes()
	s := scanner.GetErrorScanner()
	trees := errorTrees(t)
	rng := rand.New(rand.NewSource(4))
	filler := strings.Repeat("lorem ipsum, ", 25)
	perRule := max(1, iterations(t)/60)

	contexts := []func(string) string{
		func(m string) string { return m },
		func(m string) string { return m + " " + filler },
		func(m string) string { return filler + m },
		func(m string) string { return filler + m + "\r\n" + filler },
		strings.ToUpper,
		strings.ToLower,
		func(m string) string { return "<p>" + m + "</p>\n<p>" + m + "</p>" },
		func(m string) string { return "€€€" + m + "€€€" },
		func(m string) string { return m + m },
		func(m string) string { return "\u212A " + m }, // (?i) matches it for k
	}

	for i, e := range errs {
		found := 0

		// generate does not always produce a matched string: retry.
		for attempt := 0; found < perRule && attempt < perRule*maxAttempts; attempt++ {
			var sb strings.Builder

			generate(rng, trees[i], &sb)

			if !e.Regex.MatchString(sb.String()) {
				continue
			}

			found++

			for _, context := range contexts {
				checkErrors(t, s, errs, context(sb.String()))
			}
		}

		if found == 0 {
			t.Errorf("no string generated for %q", e.ErrorName)
		}
	}
}

// TestErrorScannerRandom plants random strings generated from the
// predefined regexes in random noise.
func TestErrorScannerRandom(t *testing.T) {
	errs := scanner.GetErrorRegexes()
	s := scanner.GetErrorScanner()
	trees := errorTrees(t)
	rng := rand.New(rand.NewSource(5))
	windows, matches := 0, 0

	for range iterations(t) {
		body := randomBody(rng, trees)
		checkErrors(t, s, errs, body)

		windows += countWindows(scanner.ErrorScannerEngine(s), body)
		matches += len(s.Scan(testTarget, body))
	}

	// Make sure the interesting code paths are exercised.
	if windows < iterations(t) || matches < iterations(t) {
		t.Fatalf("weak test: %d windowed scans, %d matches", windows, matches)
	}
}

// FuzzErrorScanner checks the predefined errors on random bodies.
func FuzzErrorScanner(f *testing.F) {
	rng := rand.New(rand.NewSource(6))

	for _, e := range scanner.GetErrorRegexes() {
		tree, err := syntax.Parse(e.Regex.String(), syntax.Perl)
		if err != nil {
			f.Fatal(err)
		}

		var sb strings.Builder

		generate(rng, tree.Simplify(), &sb)
		f.Add(sb.String())
	}

	errs := scanner.GetErrorRegexes()
	s := scanner.GetErrorScanner()

	f.Fuzz(func(t *testing.T, body string) {
		checkErrors(t, s, errs, body)
	})
}

// BenchmarkErrorScanner measures the errors hunting.
func BenchmarkErrorScanner(b *testing.B) {
	body := benchmarkBody()
	s := scanner.GetErrorScanner()

	b.SetBytes(int64(len(body)))

	for b.Loop() {
		s.Scan(testTarget, body)
	}
}

// BenchmarkErrorsFullScan measures the original errors hunting, for comparison.
func BenchmarkErrorsFullScan(b *testing.B) {
	body := benchmarkBody()
	errs := scanner.GetErrorRegexes()

	b.SetBytes(int64(len(body)))

	for b.Loop() {
		errorsFullScan(errs, body)
	}
}
