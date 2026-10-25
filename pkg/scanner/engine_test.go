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
	"slices"
	"strings"
	"sync"
	"testing"
	"unicode"
	"unicode/utf8"

	"github.com/edoardottt/cariddi/pkg/scanner"
)

const (
	keyRegex = `key-[a-z]{3}`
	keyText  = "key-abc"
)

// TestFoldCaseRunes checks the assumption behind the keyword variants (see
// foldVariants): the only non-ASCII runes that (?i) matches for an ASCII
// letter are the Kelvin sign (k) and the long s (s).
func TestFoldCaseRunes(t *testing.T) {
	for r := rune(0); r < utf8.RuneSelf; r++ {
		for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
			if f >= utf8.RuneSelf && f != '\u212A' && f != '\u017F' {
				t.Errorf("unexpected case folding %q -> %q", r, f)
			}
		}
	}
}

// TestEngineEdgeCases checks custom regexes on tricky texts.
func TestEngineEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		text    string
	}{
		{"empty text", keyRegex, ""},
		{"empty regex", ``, "abc"},
		{"regex matching the empty string", `(?:key)*`, "keykey x key"},
		{"keyword only", `key`, "key KEY xkeyx"},
		{"match at start and end", keyRegex, "key-abc middle key-def"},
		{"adjacent matches", keyRegex, "key-abckey-defkey-ghi"},
		{"overlapping keywords", `xoxb-[0-9]{3}`, "xoxb-xoxb-xoxb-123"},
		{"begin anchor", `^key-[a-z]{3}`, "key-abc key-def"},
		{"multiline anchors", `(?m)^key-[a-z]{3}$`, "key-abc\nkey-def\nxkey-ghi"},
		{"text anchors", `\Akey-[a-z]{3}\z`, keyText},
		{"end of text", `key-[a-z]{3}$`, "key-abc key-def"},
		{"non-ASCII literal", `(?i)ß[a-z]{3}`, "ẞabc ßdef"},
		{"only non-ASCII literal", `(?i)é+`, "éÉé x É"},
		{"4-byte case folding", `(?i)\x{10400}key[0-9]{3}`, "\U00010428KEY123 \U00010400key456"},
		{"word boundaries", `\bkey[0-9]{2}\b`, "key12 xkey34 key56x key78"},
		{"non word boundaries", `\Bkey[0-9]{2}\B`, "xkey12x key34"},
		{"lazy and greedy", `key-.+?;|key-.*`, "key-a;b key-c"},
		{"ungreedy flag", `(?U)key-[a-z]+`, keyText},
		{"unbounded with any char", `(?s)key-.*end`, "key-\n\nend"},
		{"unbounded run", `[a-z0-9.]+\.example\.com`, "foo.bar.example.com, a.example.com"},
		{"case insensitive class", `(?i)[ks]ey-[0-9]{2}`, "KEY-12 Sey-34 \u212Aey-56"},
		{"kelvin, case sensitive regex", `key-[0-9]{2}`, "\u212Aey-12 key-34"},
		{"posix class", `key-[[:alpha:]]{3}`, "key-abc key-123"},
		{"invalid utf8 in match", `key.{3}`, "key\xff\xfe\xfd key\xe2\x82"},
		{"nul bytes", `key-[a-z\x00]{3}`, "key-a\x00b \x00key-abc"},
		{"window edge in a rune", `.{0,3}key`, "x€€€€€€key"},
		{"big repeat", `key[0-9]{0,1000}`, "key" + strings.Repeat("1", 1500)},
		{"branch without keyword", `key-[0-9]+|[0-9]{5}`, "12345 key-1"},
		{"one byte keyword", `[a-z]+@[a-z]+\.com`, "mail a@b.com, x@@y.com"},
		// (?i)k also matches the Kelvin sign: the run of matchable bytes must
		// include it, also outside the keyword.
		{"case folding in a run", `(?i)k+ey-[0-9]+`, "x \u212A\u212Akey-123 \u017F"},
		// \b must see the byte before a run of matchable bytes (X is a word byte).
		{"word boundary before run", `\b[a-z.]+\.example\.com`, "Xfoo.example.com 1bar.example.com x.example.com"},
		// (?i)ß also matches the 3 bytes ẞ: the length bound must account for it.
		{"wider case folding", `(?i)ßßßßßßßßkey`, "xxxxxxxxxxẞẞẞẞẞẞẞẞkey ßßßßßßßßKEY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexes := mustCompile(t, tt.pattern)
			checkEngine(t, scanner.NewEngine(regexes), regexes, tt.text)
		})
	}
}

// TestEngineFoldVariants checks keywords written with the Kelvin sign or
// the long s, which (?i) matches for k and s: the regexes still run on
// windows only.
func TestEngineFoldVariants(t *testing.T) {
	regexes := mustCompile(t, `(?i)key-[0-9]{3}`, `(?i)secret_[a-z]{4}`, `key-[0-9]{3}`)
	e := scanner.NewEngine(regexes)
	filler := strings.Repeat("lorem ipsum ", 1000)
	text := filler + "\u212Aey-123 " + filler + "KEY-456 \u212AEY-789 \u017Fecret_abcd SECRET_EFGH " + filler

	for i, mode := range scanner.EngineModes(e, text) {
		if mode != modeWindows {
			t.Errorf("regex %q: mode %s, want %s", regexes[i], mode, modeWindows)
		}
	}

	checkEngine(t, e, regexes, text)
}

// TestEngineTooManyVariants checks that a keyword with too many k and s
// (too many variants) makes its regex run on the whole text.
func TestEngineTooManyVariants(t *testing.T) {
	regexes := mustCompile(t, `(?i)ksksksks[0-9]`)
	e := scanner.NewEngine(regexes)
	text := "KSKSKSKS1 \u212A\u017Fksksks2 ksksksks"

	if mode := scanner.EngineModes(e, text)[0]; mode != modeFull {
		t.Fatalf("mode %s, want %s", mode, modeFull)
	}

	checkEngine(t, e, regexes, text)
}

// TestEngineNoRegexes checks an engine without regexes.
func TestEngineNoRegexes(t *testing.T) {
	if got := scanner.NewEngine(nil).FindAll(keyText); len(got) != 0 {
		t.Fatalf("got %q, want no results", got)
	}
}

// TestEngineDuplicateRegexes checks regexes sharing all their keywords.
func TestEngineDuplicateRegexes(t *testing.T) {
	regexes := mustCompile(t, keyRegex, keyRegex, `(?i)KEY-[a-z]{3}`)
	checkEngine(t, scanner.NewEngine(regexes), regexes, "key-abc KEY-def")
}

// TestEngineHitsCap checks both sides of the cap on keyword occurrences:
// below it the regex runs on windows, above it on the whole text.
func TestEngineHitsCap(t *testing.T) {
	regexes := mustCompile(t, `fb[0-9]{2}`)
	e := scanner.NewEngine(regexes)
	windows, full := 0, 0

	for count := 2040; count <= 2056; count++ {
		text := strings.Repeat("fbxx", count) + "fb12"
		occurrences := count + 1

		want := modeWindows
		if occurrences > scanner.MaxHits(len(text)) {
			want = modeFull
		}

		if got := scanner.EngineModes(e, text)[0]; got != want {
			t.Fatalf("%d occurrences, cap %d: got mode %s, want %s", occurrences, scanner.MaxHits(len(text)), got, want)
		}

		if want == modeWindows {
			windows++
		} else {
			full++
		}

		checkEngine(t, e, regexes, text)
	}

	if windows == 0 || full == 0 {
		t.Fatalf("boundary not crossed: %d windowed, %d full", windows, full)
	}
}

// TestEngineConcurrentUse runs one engine from many goroutines
// (meaningful with -race).
func TestEngineConcurrentUse(t *testing.T) {
	regexes := mustCompile(t, keyRegex, `[a-z]+\.example\.com`, `(?i)aws.{0,5}[0-9]{4}`)
	e := scanner.NewEngine(regexes)
	texts := []string{keyText, "a.example.com key-def", "AWS: 1234", strings.Repeat("aws key-x ", 100)}

	var wg sync.WaitGroup

	for i := range 32 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			text := texts[i%len(texts)]
			want := fullScan(regexes, text)

			for j, got := range e.FindAll(text) {
				if !slices.Equal(got, want[j]) {
					t.Errorf("regex %d on %q: got %q, want %q", j, text, got, want[j])
				}
			}
		}()
	}

	wg.Wait()
}

// TestEngineRandomRegexes checks random regexes (like the custom ones of
// a secrets file), several per engine (sharing keywords).
func TestEngineRandomRegexes(t *testing.T) {
	rng := rand.New(rand.NewSource(2))
	windows, matches := 0, 0

	for range iterations(t) {
		var patterns []string

		for len(patterns) < 3 {
			pattern := randomRegex(rng, 4)
			if _, err := regexp.Compile(pattern); err == nil {
				patterns = append(patterns, pattern)
			}
		}

		regexes := mustCompile(t, patterns...)
		e := scanner.NewEngine(regexes)

		trees := make([]*syntax.Regexp, 0, len(patterns))
		for _, pattern := range patterns {
			trees = append(trees, mustParse(t, pattern))
		}

		for range 5 {
			text := randomBody(rng, trees)
			checkEngine(t, e, regexes, text)

			windows += countWindows(e, text)

			for _, m := range e.FindAll(text) {
				matches += len(m)
			}
		}
	}

	// Make sure the interesting code paths are exercised.
	if windows < iterations(t) || matches < iterations(t) {
		t.Fatalf("weak test: %d windowed scans, %d matches", windows, matches)
	}
}

// FuzzEngine checks random regexes on random texts.
func FuzzEngine(f *testing.F) {
	f.Add(`(?i)key-[a-z0-9]{4}\b`, "KEY-ab12 key-xyz9")
	f.Add(`[a-z.]+\.example\.com`, "a.b.example.com")
	f.Add(`\bfb(.{0,5})?[0-9]{3}$`, "fb: 123")
	f.Add(`[a-z]+@[a-z]+`, "a@b c@@d")

	f.Fuzz(func(t *testing.T, pattern, text string) {
		re, err := regexp.Compile(pattern)
		if err != nil || len(pattern) > 100 {
			return
		}

		regexes := []*regexp.Regexp{re}
		checkEngine(t, scanner.NewEngine(regexes), regexes, text)
	})
}
