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

// This file exposes internals to the tests (package scanner_test).

// EngineKeywords returns the keywords derived for each regex of e.
func EngineKeywords(e *Engine) [][]string {
	result := make([][]string, len(e.rules))

	for i, rule := range e.rules {
		for _, id := range rule.keywords {
			result[i] = append(result[i], e.matcher.keywords[id])
		}
	}

	return result
}

// EngineModes returns how each regex of e is run on text:
// "skip", "full" or "windows".
func EngineModes(e *Engine, text string) []string {
	hits, truncated := e.findKeywords(text)
	modes := make([]string, len(e.rules))

	for i := range e.rules {
		switch e.rules[i].plan(hits, truncated) {
		case scanSkip:
			modes[i] = "skip"
		case scanFull:
			modes[i] = "full"
		case scanWindows:
			modes[i] = "windows"
		}
	}

	return modes
}

// MaxHits returns how many keyword occurrences are recorded at most
// for a text of n bytes.
func MaxHits(n int) int {
	return n/hitsDivisor + minHitsCap
}

// SecretScannerEngine returns the engine of s.
func SecretScannerEngine(s *SecretScanner) *Engine {
	return s.engine
}

// ErrorScannerEngine returns the engine of s.
func ErrorScannerEngine(s *ErrorScanner) *Engine {
	return s.engine
}
