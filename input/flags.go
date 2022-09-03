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

*/

package input

import (
	"flag"
)

// Input struct.
// It contains all the possible options.
type Input struct {
	Version       bool
	Delay         int
	Concurrency   int
	Help          bool
	Examples      bool
	Plain         bool
	HTML          string
	TXT           string
	Ignore        string
	IgnoreTXT     string
	Cache         bool
	Timeout       int
	Intensive     bool
	Rua           bool
	Proxy         string
	Insecure      bool
	Secrets       bool
	SecretsFile   string
	Endpoints     bool
	EndpointsFile string
	Extensions    int
	Headers       string
	HeadersFile   string
	Errors        bool
	Info          bool
	Debug         bool
	UserAgent     string
}

// ScanFlag defines all the options taken
// as input and scan them, then it returns
// an Input struct.
func ScanFlag() Input {
	versionPtr := flag.Bool("version", false, "Print the version.")
	delayPtr := flag.Int("d", 0, "Delay between a page crawled and another.")
	concurrencyPtr := flag.Int("c", 20, "Concurrency level.")
	helpPtr := flag.Bool("h", false, "Print the help.")
	examplesPtr := flag.Bool("examples", false, "Print the examples.")
	plainPtr := flag.Bool("plain", false, "Print only the results.")
	outputHTMLPtr := flag.String("oh", "", "Write the output into an HTML file.")
	outputTXTPtr := flag.String("ot", "", "Write the output into a TXT file.")
	ignorePtr := flag.String("i", "", "Ignore the URL containing at least one of the elements of this array.")
	ignoreTXTPtr := flag.String("it", "", "Ignore the URL containing at least one of the lines of this file.")
	cachePtr := flag.Bool("cache", false, "Use the .cariddi_cache folder as cache.")
	timeoutPtr := flag.Int("t", 10, "Set timeout for the requests.")
	intensivePtr := flag.Bool("intensive", false, "Crawl searching for resources matching 2nd level domain.")
	ruaPtr := flag.Bool("rua", false, "Use a random browser user agent on every request.")
	proxyPtr := flag.String("proxy", "", "Set a Proxy to be used (http and socks5 supported).")
	insecurePtr := flag.Bool("insecure", false, "Ignore invalid HTTPS certificates")

	secretsPtr := flag.Bool("s", false, "Hunt for secrets.")
	secretsFilePtr := flag.String("sf", "", "Use an external file (txt, one per line)"+
		" to use custom regexes for secrets hunting.")

	endpointsPtr := flag.Bool("e", false, "Hunt for juicy endpoints.")
	endpointsFilePtr := flag.String("ef", "", "Use an external file (txt, one per line)"+
		" to use custom parameters for endpoints hunting.")

	extensionsPtr := flag.Int("ext", 0, "Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).")

	headersPtr := flag.String("headers", "", "Use custom headers for each request "+
		"E.g. -headers \"Cookie: auth=yes;;Client: type=2\".")
	headersFilePtr := flag.String("headersfile", "", "Read from an external file "+
		"custom headers (same format of headers flag).")

	errorsPtr := flag.Bool("err", false, "Hunt for errors in websites.")

	infoPtr := flag.Bool("info", false, "Hunt for useful informations in websites.")

	debugPtr := flag.Bool("debug", false, "Print debug information while crawling.")

	userAgentPtr := flag.String("ua", "", "Use a custom User Agent.")

	flag.Parse()

	result := Input{
		*versionPtr,
		*delayPtr,
		*concurrencyPtr,
		*helpPtr,
		*examplesPtr,
		*plainPtr,
		*outputHTMLPtr,
		*outputTXTPtr,
		*ignorePtr,
		*ignoreTXTPtr,
		*cachePtr,
		*timeoutPtr,
		*intensivePtr,
		*ruaPtr,
		*proxyPtr,
		*insecurePtr,
		*secretsPtr,
		*secretsFilePtr,
		*endpointsPtr,
		*endpointsFilePtr,
		*extensionsPtr,
		*headersPtr,
		*headersFilePtr,
		*errorsPtr,
		*infoPtr,
		*debugPtr,
		*userAgentPtr,
	}

	return result
}
