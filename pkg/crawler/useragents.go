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

* Disclaimer *: Code partially taken from:
https://github.com/gocolly/colly/blob/v1.2.0/extensions/random_user_agent.go

*/

package crawler

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	hundred = 100
)

// genOsString generates a random OS string for a User Agent.
func genOsString(osStrings []string) string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	return osStrings[rng.Intn(len(osStrings))]
}

// genFirefoxUA generates a random Firefox User Agent.
func genFirefoxUA() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Operating systems (Firefox caps macOS at 10.15).
	var OsStrings = []string{
		"Macintosh; Intel Mac OS X 10.15",
		"Windows NT 10.0; Win64; x64",
		"X11; Linux x86_64",
	}

	// Firefox versions.
	var FirefoxVersions = []float32{
		159.0,
		158.0,
		157.0,
		156.0,
		155.0,
		154.0,
		153.0,
		152.0,
		151.0,
		150.0,
		149.0,
		148.0,
		147.0,
		146.0,
		145.0,
		144.0,
		143.0,
		142.0,
		141.0,
		140.0,
		139.0,
		138.0,
		137.0,
		136.0,
		135.0,
		134.0,
		133.0,
	}

	version := FirefoxVersions[rng.Intn(len(FirefoxVersions))]

	return fmt.Sprintf("Mozilla/5.0 (%s; rv:%.1f) Gecko/20100101 Firefox/%.1f", genOsString(OsStrings), version, version)
}

// genChromeUA generates a random Chrome User Agent.
func genChromeUA() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Operating systems (Chrome freezes the OS version).
	var OsStrings = []string{
		"Macintosh; Intel Mac OS X 10_15_7",
		"Windows NT 10.0; Win64; x64",
		"X11; Linux x86_64",
	}

	// Chrome major versions (Chrome reports minor, build and patch as 0).
	var ChromeVersions = []int{
		157,
		156,
		155,
		154,
		153,
		152,
		151,
		150,
		149,
		148,
		147,
		146,
		145,
		144,
		143,
		142,
		141,
		140,
		139,
		138,
		137,
		136,
		135,
		134,
		133,
		132,
		131,
	}

	version := ChromeVersions[rng.Intn(len(ChromeVersions))]

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%d.0.0.0 Safari/537.36",
		genOsString(OsStrings), version)
}

// GenerateRandomUserAgent generates a random user agent
// (can be Chrome or Firefox).
func GenerateRandomUserAgent() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	decision := rng.Intn(hundred)

	var ua string
	if decision%2 == 0 {
		ua = genChromeUA()
	} else {
		ua = genFirefoxUA()
	}

	return ua
}
