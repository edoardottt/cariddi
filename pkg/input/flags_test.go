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

package input_test

import (
	"flag"
	"io"
	"os"
	"testing"

	"github.com/edoardottt/cariddi/pkg/input"
)

// parseRps runs ScanFlag with the given command line arguments and
// returns the resulting Rps field. Because ScanFlag binds to the global
// flag.CommandLine, the command line is reset before each invocation so
// that flags are never registered twice.
func parseRps(t *testing.T, args ...string) uint {
	t.Helper()

	os.Args = append([]string{"cariddi"}, args...)
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)

	return input.ScanFlag().Rps
}

func TestScanFlagRps(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want uint
	}{
		{
			name: "default is no rate limiting",
			args: nil,
			want: 0,
		},
		{
			name: "single digit rps",
			args: []string{"-rps", "5"},
			want: 5,
		},
		{
			name: "multi digit rps",
			args: []string{"-rps", "500"},
			want: 500,
		},
		{
			name: "explicit zero rps",
			args: []string{"-rps", "0"},
			want: 0,
		},
		{
			name: "invalid rps falls back to zero",
			args: []string{"-rps", "abc"},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseRps(t, tt.args...); got != tt.want {
				t.Errorf("Rps = %d, want %d", got, tt.want)
			}
		})
	}
}
