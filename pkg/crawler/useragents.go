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

const (
	maxRandomValue = 100
)

// genOsString generates a random OS string for a User Agent.
func genOsString() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Operating system.
	var OsStrings = []string{
		"Macintosh; Intel Mac OS X 10_10",
		"Windows NT 10.0",
		"Windows NT 5.1",
		"Windows NT 6.1; WOW64",
		"Windows NT 6.1; Win64; x64",
		"X11; Linux x86_64",
	}

	return OsStrings[rng.Intn(len(OsStrings))]
}

// genFirefoxUA generates a random Firefox User Agent.
func genFirefoxUA() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Firefox versions.
	var FirefoxVersions = []float32{
		127.0,
		126.0,
		124.0,
		123.0,
		122.0,
		121.0,
		120.0,
		119.0,
		118.0,
		117.0,
		116.0,
		115.0,
		114.0,
		113.0,
		112.0,
		111.0,
		110.0,
		109.0,
		108.0,
		107.0,
		106.0,
		105.0,
		104.0,
		103.0,
		102.0,
		101.0,
		100.0,
		58.0,
		57.0,
		56.0,
		52.0,
		48.0,
		40.0,
		35.0,
	}

	version := FirefoxVersions[rng.Intn(len(FirefoxVersions))]

	return fmt.Sprintf("Mozilla/5.0 (%s; rv:%.1f) Gecko/20100101 Firefox/%.1f", genOsString(), version, version)
}

// genChromeUA generates a random Chrome User Agent.
func genChromeUA() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Chrome versions.
	var ChromeVersions = []string{
		"126.0.6478.126",
		"124.0.6367.60",
		"123.0.6312.105",
		"121.0.6167.160",
		"120.0.6099.199",
		"119.0.6045.199",
		"118.0.5993.70",
		"117.0.5938.149",
		"116.0.5845.140",
		"115.0.5790.170",
		"114.0.5735.90",
		"113.0.5672.126",
		"112.0.5615.137",
		"111.0.5563.110",
		"109.0.5414.119",
		"108.0.5359.94",
		"107.0.5304.62",
		"106.0.5249.15",
		"105.0.5195.51",
		"104.0.5112.105",
		"103.0.5060.2",
		"102.0.5005.156",
		"101.0.4951.69",
		"100.0.4896.163",
		"65.0.3325.146",
		"64.0.3282.0",
		"41.0.2228.0",
		"40.0.2214.93",
		"37.0.2062.124",
	}

	version := ChromeVersions[rng.Intn(len(ChromeVersions))]

	return fmt.Sprintf("Mozilla/5.0 (%s) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36",
		genOsString(), version)
}

// GenerateRandomUserAgent generates a random user agent
// (can be Chrome or Firefox).
func GenerateRandomUserAgent() string {
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	decision := rng.Intn(maxRandomValue)

	var ua string
	if decision%2 == 0 {
		ua = genChromeUA()
	} else {
		ua = genFirefoxUA()
	}

	return ua
}
