/*
==========
Cariddi
==========

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
at your option) any later version.

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

func TestScanFlagRateLimitAliases(t *testing.T) {
	originalCommandLine := flag.CommandLine
	originalArgs := os.Args

	t.Cleanup(func() {
		flag.CommandLine = originalCommandLine
		os.Args = originalArgs
	})

	for _, name := range []string{"-rl", "-rate-limit"} {
		t.Run(name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet("cariddi", flag.ContinueOnError)
			flag.CommandLine.SetOutput(io.Discard)

			os.Args = []string{"cariddi", name, "7"}

			flags := input.ScanFlag()
			if flags.RateLimit != 7 {
				t.Fatalf("%s set rate limit to %d, want 7", name, flags.RateLimit)
			}
		})
	}
}
