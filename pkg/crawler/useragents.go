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

	@Author:      edoardottt, https://www.edoardoottavianelli.it

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

// genOsString generates a random OS string for a User Agent.
func genOsString() string {
	rand.Seed(time.Now().UnixNano())
	// Operating system.
	var OsStrings = []string{
		"Macintosh; Intel Mac OS X 10_10",
		"Windows NT 10.0",
		"Windows NT 5.1",
		"Windows NT 6.1; WOW64",
		"Windows NT 6.1; Win64; x64",
		"X11; Linux x86_64",
	}

	return OsStrings[rand.Intn(len(OsStrings))]
}

// genFirefoxUA generates a random Firefox User Agent.
func genFirefoxUA() string {
	rand.Seed(time.Now().UnixNano())

	// Firefox versions.
	var FirefoxVersions = []float32{
		58.0,
		57.0,
		56.0,
		52.0,
		48.0,
		40.0,
		35.0,
	}

	version := FirefoxVersions[rand.Intn(len(FirefoxVersions))]

	return fmt.Sprintf("Mozilla/5.0 (%s; rv:%.1f) Gecko/20100101 Firefox/%.1f", genOsString(), version, version)
}

// genChromeUA generates a random Chrome User Agent.
func genChromeUA() string {
	rand.Seed(time.Now().UnixNano())

	// Chrome versions.
	var ChromeVersions = []string{
		"65.0.3325.146",
		"64.0.3282.0",
		"41.0.2228.0",
		"40.0.2214.93",
		"37.0.2062.124",
	}

	version := ChromeVersions[rand.Intn(len(ChromeVersions))]

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		genOsString(), version)
}

// GenerateRandomUserAgent generates a random user agent
// (can be Chrome or Firefox).
func GenerateRandomUserAgent() string {
	rand.Seed(time.Now().UnixNano())

	decision := rand.Intn(100)

	var ua string
	if decision%2 == 0 {
		ua = genChromeUA()
	} else {
		ua = genFirefoxUA()
	}

	return ua
}
