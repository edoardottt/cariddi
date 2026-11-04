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

package input

import "io"

// This file exposes internals to the tests (package input_test).

// MaxTargetLength is the maximum length of an input line (see maxTargetLength).
const MaxTargetLength = maxTargetLength

// ParseHeaders returns the headers contained in input (see parseHeaders).
func ParseHeaders(input string) map[string]string {
	return parseHeaders(input)
}

// ScanFlagArgs scans the flags in args (see scanFlag).
func ScanFlagArgs(args []string) (Input, []string) {
	return scanFlag(args)
}

// ReadTargets returns the targets read from r (see readTargets).
func ReadTargets(r io.Reader) []string {
	return readTargets(r, "the test input", false)
}
