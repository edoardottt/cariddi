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
	"container/heap"
	"iter"
	"regexp"
	"sync"
	"unicode/utf8"
)

const (
	// Keyword occurrences recorded per text are capped to
	// len(text)/hitsDivisor + minHitsCap (rules missing some
	// occurrences fall back to a full scan).
	hitsDivisor = 8
	minHitsCap  = 1024
)

// Engine finds the matches of a list of regular expressions in texts.
//
// FindAll returns exactly what regexp.FindAllString(text, -1) returns for
// each regex, but it skips most of the work:
//  1. every match of a regex contains one of its keywords (derived from the
//     regex itself, e.g. "adafruit", ignoring case): if none is in the text,
//     the regex can't match, so it is skipped;
//  2. otherwise, matches can only be near the keywords found (within the
//     maximum match length, or within the run of bytes the regex can match),
//     so the regex only runs on those small windows of the text.
//
// Regexes without keywords are run on the whole text, as usual.
// The regexes must be compiled with regexp.Compile or regexp.MustCompile.
// An Engine is safe for concurrent use.
type Engine struct {
	rules   []engineRule
	matcher *keywordMatcher
}

type engineRule struct {
	re       *regexp.Regexp
	keywords []int32   // ids of the keywords of the rule, nil if unknown.
	maxLen   int       // upper bound on the match length in bytes, -1 if unbounded.
	alphabet *alphabet // bytes a match can contain, nil if unknown or any.
}

type occurrence struct {
	start, end int
}

type window struct {
	lo, hi int
}

type scanMode int

const (
	scanSkip scanMode = iota
	scanFull
	scanWindows
)

// NewEngine builds an engine for the given regexes.
func NewEngine(regexes []*regexp.Regexp) *Engine {
	e := &Engine{rules: make([]engineRule, len(regexes))}
	ids := map[string]int32{}

	var keywords []string

	for i, re := range regexes {
		rule := engineRule{re: re, maxLen: -1}

		tree, err := parseRegex(re.String())
		if err == nil {
			for _, keyword := range requiredKeywords(tree) {
				variants := foldVariants(shortenKeyword(keyword))
				if variants == nil {
					rule.keywords = nil // too many variants: always run the regex.

					break
				}

				for _, variant := range variants {
					id, ok := ids[variant]
					if !ok {
						id = int32(len(keywords))
						ids[variant] = id
						keywords = append(keywords, variant)
					}

					rule.keywords = append(rule.keywords, id)
				}
			}

			if !hasTextAnchors(tree) {
				rule.maxLen = maxMatchLen(tree)
				if rule.maxLen < 0 {
					rule.alphabet = newAlphabet(tree)
				}
			}
		}

		e.rules[i] = rule
	}

	e.matcher = newKeywordMatcher(keywords)

	return e
}

// FindAll returns, for each regex (in the order given to NewEngine),
// what its FindAllString(text, -1) returns.
func (e *Engine) FindAll(text string) [][]string {
	hits, truncated := e.findKeywords(text)
	results := make([][]string, len(e.rules))

	var wg sync.WaitGroup

	for i := range e.rules {
		rule := &e.rules[i]

		mode := rule.plan(hits, truncated)
		if mode == scanSkip {
			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			if mode == scanFull {
				results[i] = rule.re.FindAllString(text, -1)
			} else {
				results[i] = rule.findInWindows(text, hits, e.matcher.keywords)
			}
		}()
	}

	wg.Wait()

	return results
}

// findKeywords finds the keyword occurrences in text.
func (e *Engine) findKeywords(text string) (hits [][]int, truncated []bool) {
	return e.matcher.find(text, len(text)/hitsDivisor+minHitsCap)
}

// plan decides how the rule must be run on a text.
func (r *engineRule) plan(hits [][]int, truncated []bool) scanMode {
	if r.keywords == nil {
		return scanFull
	}

	found := false

	for _, id := range r.keywords {
		if truncated[id] {
			return scanFull
		}

		found = found || len(hits[id]) > 0
	}

	switch {
	case !found:
		return scanSkip
	case r.maxLen < 0 && r.alphabet == nil:
		return scanFull
	default:
		return scanWindows
	}
}

// findInWindows returns the same matches as FindAllString on the whole
// text, running the regex only on windows around the keyword occurrences.
//
// Every match contains a keyword occurrence. For a bounded rule (every match
// at most maxLen bytes) a match containing an occurrence at p lies within
// [p-maxLen, p+maxLen], and so does everything the regex engine reads to
// find it, plus one rune of look-ahead. For an unbounded rule, a match can't
// contain bytes outside the rule alphabet, so it lies within the run of
// matchable bytes around p; the window also includes the delimiting byte on
// each side, so the engine sees the same context as in the full text.
// Windows are aligned to rune boundaries and merged when they overlap.
//
// The occurrences are visited in order and every window is scanned as soon
// as the next occurrence can't extend it: nothing is copied or stored, so
// the memory doesn't depend on how many occurrences the text contains.
func (r *engineRule) findInWindows(text string, hits [][]int, keywords []string) []string {
	var (
		matches []string
		cur     window
		open    bool
	)

	for o := range r.occurrences(hits, keywords) {
		// Occurrences inside the current run window have the same window.
		if r.maxLen < 0 && open && o.end <= cur.hi {
			continue
		}

		var w window
		if r.maxLen >= 0 {
			w = window{max(0, o.start-r.maxLen), min(len(text), o.start+r.maxLen+1)}
		} else {
			w = r.runWindow(text, o)
		}

		w = alignWindow(text, w)

		if open && w.lo <= cur.hi {
			cur.hi = max(cur.hi, w.hi)

			continue
		}

		if open {
			matches = r.appendMatches(matches, text, cur)
		}

		cur, open = w, true
	}

	if open {
		matches = r.appendMatches(matches, text, cur)
	}

	return matches
}

// appendMatches appends to matches the matches of the rule in the window w.
func (r *engineRule) appendMatches(matches []string, text string, w window) []string {
	part := text[w.lo:w.hi]
	for _, loc := range r.re.FindAllStringIndex(part, -1) {
		matches = append(matches, part[loc[0]:loc[1]])
	}

	return matches
}

// occurrences yields the occurrences of the rule keywords in text, sorted
// by start: it merges the hits of each keyword (already sorted).
func (r *engineRule) occurrences(hits [][]int, keywords []string) iter.Seq[occurrence] {
	return func(yield func(occurrence) bool) {
		cursors := make(hitHeap, 0, len(r.keywords))

		for _, id := range r.keywords {
			if len(hits[id]) > 0 {
				cursors = append(cursors, hitCursor{starts: hits[id], length: len(keywords[id])})
			}
		}

		heap.Init(&cursors)

		for len(cursors) > 0 {
			c := &cursors[0]
			if !yield(occurrence{c.starts[0], c.starts[0] + c.length}) {
				return
			}

			if c.starts = c.starts[1:]; len(c.starts) == 0 {
				heap.Pop(&cursors)
			} else {
				heap.Fix(&cursors, 0)
			}
		}
	}
}

// hitCursor points to the next occurrences of a keyword.
type hitCursor struct {
	starts []int // start offsets, in increasing order.
	length int   // keyword length.
}

// hitHeap is a min-heap of cursors, by next start (see container/heap).
type hitHeap []hitCursor

func (h hitHeap) Len() int           { return len(h) }
func (h hitHeap) Less(i, j int) bool { return h[i].starts[0] < h[j].starts[0] }
func (h hitHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *hitHeap) Push(x any) { *h = append(*h, x.(hitCursor)) }

func (h *hitHeap) Pop() any {
	old := *h
	c := old[len(old)-1]
	*h = old[:len(old)-1]

	return c
}

// runWindow returns the run of matchable bytes around o, plus the
// delimiting byte on each side.
func (r *engineRule) runWindow(text string, o occurrence) window {
	lo := o.start
	for lo > 0 && !r.alphabet.excluded(text[lo-1]) {
		lo--
	}

	if lo > 0 {
		lo--
	}

	hi := o.end
	for hi < len(text) && !r.alphabet.excluded(text[hi]) {
		hi++
	}

	if hi < len(text) {
		hi++
	}

	return window{lo, hi}
}

// alignWindow moves the window bounds to rune boundaries, so that the
// regex engine decodes the window exactly as it decodes the full text.
// This is only a precaution: a match can't start or end at a window bound.
func alignWindow(text string, w window) window {
	for w.lo > 0 && !utf8.RuneStart(text[w.lo]) {
		w.lo--
	}

	for w.hi < len(text) && !utf8.RuneStart(text[w.hi]) {
		w.hi++
	}

	return w
}
