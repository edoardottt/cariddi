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

import "strings"

const (
	alphabetSize = 256
	caseOffset   = 'a' - 'A'
	// maxKeywordLen is the maximum keyword length: longer keywords are
	// shortened to one of their substrings (which every match contains too),
	// to keep the automaton small.
	maxKeywordLen = 16
	// maxFoldVariants caps the variants of a keyword (see foldVariants).
	maxFoldVariants = 64
)

// shortenKeyword returns the substring of keyword of at most maxKeywordLen
// bytes with the fewest k and s (each one doubles the variants).
func shortenKeyword(keyword string) string {
	if len(keyword) <= maxKeywordLen {
		return keyword
	}

	best, bestCount := "", maxKeywordLen+1

	for i := 0; i+maxKeywordLen <= len(keyword); i++ {
		sub := keyword[i : i+maxKeywordLen]
		if count := strings.Count(sub, "k") + strings.Count(sub, "s"); count < bestCount {
			best, bestCount = sub, count
		}
	}

	return best
}

// foldVariants returns keyword and its variants with k and s replaced by the
// Kelvin sign and the long s: (?i) matches them too, but the keyword search
// only ignores ASCII case. They are the only non-ASCII runes (?i) matches
// for an ASCII letter. It returns nil if there are more than maxFoldVariants.
func foldVariants(keyword string) []string {
	variants := []string{""}

	for i := range len(keyword) {
		alternatives := []string{keyword[i : i+1]}

		switch keyword[i] {
		case 'k':
			alternatives = append(alternatives, "\u212A")
		case 's':
			alternatives = append(alternatives, "\u017F")
		}

		if len(variants)*len(alternatives) > maxFoldVariants {
			return nil
		}

		next := make([]string, 0, len(variants)*len(alternatives))

		for _, variant := range variants {
			for _, alternative := range alternatives {
				next = append(next, variant+alternative)
			}
		}

		variants = next
	}

	return variants
}

// keywordMatcher finds all the occurrences of a set of lowercase ASCII
// keywords in a text, ignoring ASCII case, in a single pass
// (Aho-Corasick automaton with a full transition table).
type keywordMatcher struct {
	next     [][alphabetSize]int32 // next[state][byte] = next state.
	outputs  [][]int32             // keywords ending in each state.
	keywords []string
}

// newKeywordMatcher builds the automaton for the given lowercase keywords
// (keyword ids are their indexes).
func newKeywordMatcher(keywords []string) *keywordMatcher {
	m := &keywordMatcher{
		next:     make([][alphabetSize]int32, 1),
		outputs:  make([][]int32, 1),
		keywords: keywords,
	}

	// Build the trie.
	for id, keyword := range keywords {
		state := int32(0)

		for i := range len(keyword) {
			c := keyword[i]
			if m.next[state][c] == 0 {
				m.next = append(m.next, [alphabetSize]int32{})
				m.outputs = append(m.outputs, nil)
				m.next[state][c] = int32(len(m.next) - 1)
			}

			state = m.next[state][c]
		}

		m.outputs[state] = append(m.outputs[state], int32(id))
	}

	// Breadth-first: compute failure links and complete the transitions.
	fail := make([]int32, len(m.next))
	queue := make([]int32, 0, len(m.next))

	for c := range alphabetSize {
		if t := m.next[0][c]; t != 0 {
			queue = append(queue, t)
		}
	}

	for len(queue) > 0 {
		state := queue[0]
		queue = queue[1:]

		m.outputs[state] = append(m.outputs[state], m.outputs[fail[state]]...)

		for c := range alphabetSize {
			if t := m.next[state][c]; t != 0 {
				fail[t] = m.next[fail[state]][c]
				queue = append(queue, t)
			} else {
				m.next[state][c] = m.next[fail[state]][c]
			}
		}
	}

	// Keywords are lowercase: make uppercase letters behave the same.
	for state := range m.next {
		for c := 'A'; c <= 'Z'; c++ {
			m.next[state][c] = m.next[state][c+caseOffset]
		}
	}

	return m
}

// find returns the start offsets of the occurrences of each keyword in
// text, in increasing order. It records at most maxHits offsets in total:
// truncated[id] is true if some occurrences of keyword id were not recorded.
func (m *keywordMatcher) find(text string, maxHits int) (hits [][]int, truncated []bool) {
	hits = make([][]int, len(m.keywords))
	truncated = make([]bool, len(m.keywords))
	recorded := 0
	state := int32(0)

	for i := range len(text) {
		state = m.next[state][text[i]]

		for _, id := range m.outputs[state] {
			if recorded == maxHits {
				truncated[id] = true

				continue
			}

			hits[id] = append(hits[id], i+1-len(m.keywords[id]))
			recorded++
		}
	}

	return hits, truncated
}
