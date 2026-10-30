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
	"strconv"
	"strings"
	"testing"

	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

const testTarget = "https://example.com/page"

// secretsFullScan is the original secrets hunting (every regex on the
// whole body, false positives removed), run sequentially in rule order.
func secretsFullScan(secrets []scanner.Secret, body string) []scanner.SecretMatched {
	var result []scanner.SecretMatched

	for _, secret := range secrets {
		for _, match := range secret.Regex.FindAllStringSubmatch(body, -1) {
			isFalsePositive := false

			for _, fp := range secret.FalsePositives {
				if strings.Contains(strings.ToLower(match[0]), fp) {
					isFalsePositive = true

					break
				}
			}

			if !isFalsePositive {
				result = append(result, scanner.SecretMatched{Secret: secret, URL: testTarget, Match: match[0]})
			}
		}
	}

	return result
}

// checkSecrets fails if s and the original secrets hunting find different secrets in body.
func checkSecrets(t *testing.T, s *scanner.SecretScanner, secrets []scanner.Secret, body string) {
	t.Helper()

	want := secretsFullScan(secrets, body)
	got := s.Scan(testTarget, body)

	same := len(want) == len(got)
	for i := 0; same && i < len(want); i++ {
		w, g := want[i], got[i]
		same = w.Match == g.Match && w.URL == g.URL && w.Secret.Name == g.Secret.Name &&
			w.Secret.Description == g.Secret.Description && w.Secret.Poc == g.Secret.Poc &&
			w.Secret.Regex.String() == g.Secret.Regex.String() &&
			slices.Equal(w.Secret.FalsePositives, g.Secret.FalsePositives)
	}

	if !same {
		t.Fatalf("body %q:\nwant %v\ngot  %v", body, secretNames(want), secretNames(got))
	}
}

// secretNames describes matches, for test failures.
func secretNames(matches []scanner.SecretMatched) []string {
	result := make([]string, 0, len(matches))
	for _, m := range matches {
		result = append(result, m.Secret.Name+": "+strconv.Quote(m.Match))
	}

	return result
}

// TestBuiltinSecretsHaveKeywords checks that no predefined secret regex runs on every body.
func TestBuiltinSecretsHaveKeywords(t *testing.T) {
	keywords := scanner.EngineKeywords(scanner.SecretScannerEngine(scanner.GetSecretScanner()))

	for i, secret := range scanner.GetSecretRegexes() {
		if len(keywords[i]) == 0 {
			t.Errorf("no keywords derived for %q: it will be run on every body, add a literal to the regex", secret.Name)
		}
	}
}

// TestSecretScannerEdgeCases checks the predefined secrets on tricky bodies.
func TestSecretScannerEdgeCases(t *testing.T) {
	secrets := scanner.GetSecretRegexes()
	s := scanner.GetSecretScanner()
	adafruit := "adafruit_key = \"" + strings.Repeat("a1", 16)
	facebook := "'0123456789abcdef0123456789abcdef'"

	bodies := map[string]string{
		"empty":                     "",
		"keyword at start":          "AKIAABCDEFGHIJKLMNOP is a key",
		"secret at end of text":     "var x; " + adafruit,
		"secret then delimiter":     adafruit + "\"",
		"crlf":                      adafruit + "\r\n",
		"kelvin sign":               "x = 'A\u212AIA1234567890ABCDEF';",
		"long s":                    "token: \u017Fk-ant-api01-" + strings.Repeat("a", 95),
		"invalid utf8 before":       "\xff\xfeAKIAABCDEFGHIJKLMNOP\xe2\x84",
		"multibyte near keyword":    "€€€aws€\"" + strings.Repeat("A", 40) + "\"€€€",
		"dense keywords (cap)":      strings.Repeat("fb", 3000) + " facebook = " + facebook,
		"long run (unbounded rule)": strings.Repeat("a.", 1000) + "s3.amazonaws.com/bucket " + strings.Repeat("x-", 100),
		"same match, two rules":     "token ghs_" + strings.Repeat("A", 36) + " end",
		"overlapping keywords":      "xoxb-xoxb-xoxb-1234567890abc AKIAAKIAAKIAAKIAAKIAAKIA",
		"uppercase keywords":        "STRIPE_KEY=SK_LIVE_" + strings.Repeat("A", 24) + " sk_live_" + strings.Repeat("b", 24),
		"false positive":            "https://facebook.com/x " + facebook + " fb " + facebook,
		"binary":                    "\x00\x01\x02AKIAABCDEFGHIJKLMNOP\x00\xff",
		"utf-16 is not decoded":     "A\x00K\x00I\x00A\x00B\x00C\x00D\x00E\x00F\x00G\x00H\x00I\x00J\x00K\x00",
		"private key":               "-----BEGIN RSA PRIVATE KEY-----\nMIIEow...",
		"url secrets":               "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX",
	}

	for name, body := range bodies {
		t.Run(name, func(t *testing.T) {
			checkSecrets(t, s, secrets, body)
		})
	}
}

// TestSecretScannerCustomSecrets checks names, false positives and an empty list.
func TestSecretScannerCustomSecrets(t *testing.T) {
	secrets := []scanner.Secret{
		{Name: "a", Regex: *regexp.MustCompile(`key-[a-z]{4}`), FalsePositives: []string{"test"}},
		{Name: "b", Regex: *regexp.MustCompile(`(?i)KEY-[a-z]{4}`), Description: "d", Poc: "p"},
	}
	s := scanner.NewSecretScanner(secrets)

	checkSecrets(t, s, secrets, "key-abcd key-test KEY-wxyz")

	if got := scanner.NewSecretScanner(nil).Scan(testTarget, "key-abcd"); len(got) != 0 {
		t.Fatalf("empty scanner found %v", secretNames(got))
	}
}

// TestSecretScannerRandom plants random strings generated from the
// predefined regexes in random noise.
func TestSecretScannerRandom(t *testing.T) {
	secrets := scanner.GetSecretRegexes()
	s := scanner.GetSecretScanner()
	rng := rand.New(rand.NewSource(1))

	trees := make([]*syntax.Regexp, 0, len(secrets))
	for _, secret := range secrets {
		trees = append(trees, mustParse(t, secret.Regex.String()))
	}

	windows, matches := 0, 0

	for range iterations(t) {
		body := randomBody(rng, trees)
		checkSecrets(t, s, secrets, body)

		windows += countWindows(scanner.SecretScannerEngine(s), body)
		matches += len(s.Scan(testTarget, body))
	}

	// Make sure the interesting code paths are exercised.
	if windows < iterations(t) || matches < iterations(t) {
		t.Fatalf("weak test: %d windowed scans, %d matches", windows, matches)
	}
}

// FuzzSecretScanner checks the predefined secrets on random bodies.
func FuzzSecretScanner(f *testing.F) {
	f.Add("AKIAABCDEFGHIJKLMNOP")
	f.Add("adafruit_key = \"" + strings.Repeat("a1", 16) + "\"")
	f.Add("x = 'A\u212AIA1234567890ABCDEF';")
	f.Add("bucket.s3.amazonaws.com/x")
	f.Add("facebook = '0123456789abcdef0123456789abcdef' fb")

	secrets := scanner.GetSecretRegexes()
	s := scanner.GetSecretScanner()

	f.Fuzz(func(t *testing.T, body string) {
		checkSecrets(t, s, secrets, body)
	})
}

// benchmarkBody returns a synthetic, JavaScript-like body.
func benchmarkBody() string {
	rng := rand.New(rand.NewSource(3))
	words := []string{
		"function", "return", "this.path", "var", "const", "options", "window", "document",
		"keyCode", "value", "=", ";", "(", ")", "{", "}", "\"", ".", ",", " ", "\n", "aws.region", "fb", "data",
	}

	var sb strings.Builder
	for sb.Len() < benchmarkBodySize {
		sb.WriteString(words[rng.Intn(len(words))])
	}

	return sb.String()
}

const benchmarkBodySize = 256 * 1024

// BenchmarkSecretScanner measures the secrets hunting.
func BenchmarkSecretScanner(b *testing.B) {
	body := benchmarkBody()
	s := scanner.GetSecretScanner()

	b.SetBytes(int64(len(body)))

	for b.Loop() {
		s.Scan(testTarget, body)
	}
}

// BenchmarkSecretsFullScan measures the original secrets hunting, for comparison.
func BenchmarkSecretsFullScan(b *testing.B) {
	body := benchmarkBody()
	secrets := scanner.GetSecretRegexes()

	b.SetBytes(int64(len(body)))

	for b.Loop() {
		secretsFullScan(secrets, body)
	}
}
