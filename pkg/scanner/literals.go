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
	"regexp/syntax"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// This file analyzes the syntax tree of a regular expression to find out,
// without running it:
//   - requiredKeywords: literals such that every match contains one of them;
//   - maxMatchLen: an upper bound on the length of every match;
//   - newAlphabet: the bytes that can appear inside a match.
// Every answer is conservative: when in doubt, the analysis gives up
// (no keywords, unbounded length, every byte allowed) and the regex is
// simply run on the whole body, as before.

const (
	// minKeywordLen is the minimum keyword length. Even a one byte keyword
	// can be useful (e.g. "@" for emails): if it is too frequent, its rule
	// falls back to a full scan (see hitsDivisor).
	minKeywordLen = 1
	// maxLiteralSet caps the size of literal sets, to avoid blowups.
	maxLiteralSet = 64
	// maxClassSize is the largest character class expanded into literals.
	maxClassSize = 8
)

// literalInfo describes the strings matched by a regex node.
type literalInfo struct {
	exact  []string // the node matches exactly one of these (nil = unknown).
	prefix []string // every match starts with one of these.
	suffix []string // every match ends with one of these.
	match  []string // every match contains one of these (nil = unknown).
}

var anyString = literalInfo{prefix: []string{""}, suffix: []string{""}} //nolint: gochecknoglobals

// parseRegex parses expr exactly like regexp.Compile does.
func parseRegex(expr string) (*syntax.Regexp, error) {
	re, err := syntax.Parse(expr, syntax.Perl)
	if err != nil {
		return nil, err
	}

	return re.Simplify(), nil
}

// requiredKeywords returns a set of lowercase ASCII keywords such that
// every match of re contains at least one of them, ignoring ASCII case.
// It returns nil if no useful set can be proven.
func requiredKeywords(re *syntax.Regexp) []string {
	info := analyzeLiterals(re)

	set := bestSet(info.candidates()...)
	if shortest(set) < minKeywordLen {
		return nil
	}

	return dropSuperstrings(set)
}

// candidates returns the sets of which every match contains an element.
func (l literalInfo) candidates() [][]string {
	if l.exact != nil {
		return [][]string{l.exact}
	}

	return [][]string{l.match, l.prefix, l.suffix}
}

// prefixes returns the set of which every match starts with an element.
func (l literalInfo) prefixes() []string {
	if l.exact != nil {
		return l.exact
	}

	return l.prefix
}

// suffixes returns the set of which every match ends with an element.
func (l literalInfo) suffixes() []string {
	if l.exact != nil {
		return l.exact
	}

	return l.suffix
}

// analyzeLiterals computes, recursively, what the strings matched by re
// contain (see literalInfo).
func analyzeLiterals(re *syntax.Regexp) literalInfo {
	switch re.Op { //nolint: exhaustive // other operators are handled by default.
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText,
		syntax.OpEndText, syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return literalInfo{exact: []string{""}}
	case syntax.OpLiteral:
		acc := literalInfo{exact: []string{""}}

		for _, r := range re.Rune {
			if r >= utf8.RuneSelf {
				// Non-ASCII case folding does not map 1:1 to ASCII lowercasing.
				acc = concatLiterals(acc, anyString)
			} else {
				acc = concatLiterals(acc, literalInfo{exact: []string{string(unicode.ToLower(r))}})
			}
		}

		return acc
	case syntax.OpCharClass:
		return analyzeCharClass(re)
	case syntax.OpCapture:
		return analyzeLiterals(re.Sub[0])
	case syntax.OpQuest:
		if sub := analyzeLiterals(re.Sub[0]); sub.exact != nil && len(sub.exact) < maxLiteralSet {
			return literalInfo{exact: union(sub.exact, []string{""})}
		}

		return anyString
	case syntax.OpPlus:
		return analyzeRepeated(analyzeLiterals(re.Sub[0]))
	case syntax.OpRepeat:
		if re.Min == 0 {
			return anyString
		}

		return analyzeRepeated(analyzeLiterals(re.Sub[0]))
	case syntax.OpConcat:
		acc := literalInfo{exact: []string{""}}
		for _, sub := range re.Sub {
			acc = concatLiterals(acc, analyzeLiterals(sub))
		}

		return acc
	case syntax.OpAlternate:
		return alternateLiterals(re.Sub)
	default: // OpStar, OpAnyChar, OpAnyCharNotNL, OpNoMatch.
		return anyString
	}
}

// analyzeRepeated handles X{n,m} with n >= 1: a match starts with
// a prefix of X, ends with a suffix of X and contains what X requires.
func analyzeRepeated(sub literalInfo) literalInfo {
	return literalInfo{prefix: sub.prefixes(), suffix: sub.suffixes(), match: bestSet(sub.candidates()...)}
}

// analyzeCharClass expands a small ASCII character class into its
// (lowercase) characters, e.g. [a-c] matches exactly "a", "b" or "c".
func analyzeCharClass(re *syntax.Regexp) literalInfo {
	var chars []string

	for i := 0; i < len(re.Rune); i += 2 {
		lo, hi := re.Rune[i], re.Rune[i+1]
		if hi >= utf8.RuneSelf || int(hi-lo) >= maxClassSize*2 {
			return anyString
		}

		for r := lo; r <= hi; r++ {
			chars = union(chars, []string{string(unicode.ToLower(r))})
		}

		if len(chars) > maxClassSize {
			return anyString
		}
	}

	return literalInfo{exact: chars}
}

// concatLiterals combines two consecutive nodes, e.g. "ab" followed
// by [cd] matches exactly "abc" or "abd".
func concatLiterals(a, b literalInfo) literalInfo {
	if a.exact != nil && b.exact != nil {
		if x := crossProduct(a.exact, b.exact); x != nil {
			return literalInfo{exact: x}
		}
	}

	result := literalInfo{prefix: a.prefixes(), suffix: b.suffixes()}

	if a.exact != nil {
		if x := crossProduct(a.exact, b.prefixes()); x != nil {
			result.prefix = x
		}
	}

	if b.exact != nil {
		if x := crossProduct(a.suffixes(), b.exact); x != nil {
			result.suffix = x
		}
	}

	result.match = bestSet(a.match, b.match, crossProduct(a.suffixes(), b.prefixes()), a.exact, b.exact)

	return result
}

// alternateLiterals combines the branches of an alternation: a match
// contains what one of the branches requires.
func alternateLiterals(subs []*syntax.Regexp) literalInfo {
	var exact, prefix, suffix, match []string

	allExact, allRequire := true, true

	for _, sub := range subs {
		info := analyzeLiterals(sub)
		if info.exact == nil {
			allExact = false
		} else {
			exact = union(exact, info.exact)
		}

		prefix = capSet(union(prefix, info.prefixes()))
		suffix = capSet(union(suffix, info.suffixes()))

		// A branch without requirements means the alternation has none.
		if required := bestSet(info.candidates()...); shortest(required) < 1 {
			allRequire = false
		} else {
			match = union(match, required)
		}
	}

	if allExact && len(exact) <= maxLiteralSet {
		return literalInfo{exact: exact}
	}

	if !allRequire || len(match) > maxLiteralSet {
		match = nil
	}

	return literalInfo{prefix: prefix, suffix: suffix, match: match}
}

// crossProduct returns {x+y}, or nil if the result would be too large.
func crossProduct(a, b []string) []string {
	if a == nil || b == nil || len(a)*len(b) > maxLiteralSet {
		return nil
	}

	result := make([]string, 0, len(a)*len(b))

	for _, x := range a {
		for _, y := range b {
			result = union(result, []string{x + y})
		}
	}

	return result
}

// capSet replaces an oversized prefix/suffix set with the trivial one.
func capSet(set []string) []string {
	if len(set) > maxLiteralSet {
		return []string{""}
	}

	return set
}

// bestSet picks the most selective set: the one whose shortest string is
// the longest (ties broken by fewer elements).
func bestSet(sets ...[]string) []string {
	var best []string

	for _, set := range sets {
		if set == nil {
			continue
		}

		if best == nil || shortest(set) > shortest(best) ||
			(shortest(set) == shortest(best) && len(set) < len(best)) {
			best = set
		}
	}

	return best
}

// shortest returns the length of the shortest string in set, -1 if nil.
func shortest(set []string) int {
	if set == nil {
		return -1
	}

	result := -1

	for _, s := range set {
		if result == -1 || len(s) < result {
			result = len(s)
		}
	}

	return result
}

// union returns the strings of a and b, without duplicates.
func union(a, b []string) []string {
	result := append([]string(nil), a...)

	for _, s := range b {
		if !slices.Contains(result, s) {
			result = append(result, s)
		}
	}

	return result
}

// dropSuperstrings drops keywords containing another keyword of the set:
// a text containing the longer one always contains the shorter one.
func dropSuperstrings(set []string) []string {
	var result []string

	for i, s := range set {
		redundant := false

		for j, t := range set {
			if i != j && len(t) < len(s) && strings.Contains(s, t) {
				redundant = true

				break
			}
		}

		if !redundant {
			result = append(result, s)
		}
	}

	return result
}

// maxMatchLen returns an upper bound, in bytes, on the length of any
// match of re, or -1 if matches can be arbitrarily long.
func maxMatchLen(re *syntax.Regexp) int {
	switch re.Op { //nolint: exhaustive // other operators are handled by default.
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText,
		syntax.OpEndText, syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return 0
	case syntax.OpLiteral:
		total := 0

		for _, r := range re.Rune {
			n := utf8.RuneLen(r)

			if re.Flags&syntax.FoldCase != 0 {
				// (?i) also matches the other cases of r, which may be longer
				// (e.g. k also matches the 3 bytes Kelvin sign).
				for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
					n = max(n, utf8.RuneLen(f))
				}
			}

			total += max(n, 1)
		}

		return total
	case syntax.OpCharClass, syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		return utf8.UTFMax
	case syntax.OpCapture, syntax.OpQuest:
		return maxMatchLen(re.Sub[0])
	case syntax.OpRepeat:
		sub := maxMatchLen(re.Sub[0])
		if re.Max < 0 || sub < 0 {
			return -1
		}

		return re.Max * sub
	case syntax.OpConcat:
		total := 0

		for _, sub := range re.Sub {
			n := maxMatchLen(sub)
			if n < 0 {
				return -1
			}

			total += n
		}

		return total
	case syntax.OpAlternate:
		longest := 0

		for _, sub := range re.Sub {
			n := maxMatchLen(sub)
			if n < 0 {
				return -1
			}

			longest = max(longest, n)
		}

		return longest
	default: // OpStar, OpPlus, OpNoMatch.
		return -1
	}
}

// hasTextAnchors reports whether re contains ^, \A or a multi-line $.
// Such regexes are run on the whole text. This is only a precaution: no
// match can start at the beginning of a window or end at its end, so the
// anchors would work there too, but anchored regexes are rare.
// A plain $ (end of text) is not checked.
func hasTextAnchors(re *syntax.Regexp) bool {
	switch re.Op { //nolint: exhaustive // other operators are handled by default.
	case syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText:
		return true
	default:
		return slices.ContainsFunc(re.Sub, hasTextAnchors)
	}
}

// alphabet is the set of bytes that can appear inside a match.
type alphabet struct {
	ascii    [utf8.RuneSelf]bool
	nonASCII bool // if any non-ASCII rune can match, every byte >= 0x80 can.
}

// newAlphabet returns the bytes that can appear in a match of re, or nil
// if every byte can (then there is nothing to gain from it).
func newAlphabet(re *syntax.Regexp) *alphabet {
	a := &alphabet{}
	a.add(re)

	for c := range utf8.RuneSelf {
		if !a.ascii[c] {
			return a
		}
	}

	if !a.nonASCII {
		return a
	}

	return nil
}

// excluded reports whether byte c can never appear inside a match.
func (a *alphabet) excluded(c byte) bool {
	if c >= utf8.RuneSelf {
		return !a.nonASCII
	}

	return !a.ascii[c]
}

// addRune adds the bytes of r to the alphabet.
func (a *alphabet) addRune(r rune) {
	if r < utf8.RuneSelf {
		a.ascii[r] = true
	} else {
		a.nonASCII = true
	}
}

// add adds to the alphabet every byte that a match of re can contain.
func (a *alphabet) add(re *syntax.Regexp) {
	switch re.Op { //nolint: exhaustive // other operators are handled by default.
	case syntax.OpLiteral:
		for _, r := range re.Rune {
			a.addRune(r)

			if re.Flags&syntax.FoldCase != 0 {
				for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
					a.addRune(f)
				}
			}
		}
	case syntax.OpCharClass:
		for i := 0; i < len(re.Rune); i += 2 {
			for r := re.Rune[i]; r <= re.Rune[i+1] && r < utf8.RuneSelf; r++ {
				a.ascii[r] = true
			}

			if re.Rune[i+1] >= utf8.RuneSelf {
				a.nonASCII = true
			}
		}
	case syntax.OpAnyChar, syntax.OpAnyCharNotNL:
		for c := range utf8.RuneSelf {
			if re.Op == syntax.OpAnyChar || c != '\n' {
				a.ascii[c] = true
			}
		}

		a.nonASCII = true
	default:
		for _, sub := range re.Sub {
			a.add(sub)
		}
	}
}
