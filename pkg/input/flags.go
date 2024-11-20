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

	@Author:      edoardottt, https://www.edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package input

import (
	"flag"
)

const (
	DefaultConcurrency = 20
	TimeoutRequest     = 10
)

// Input struct.
// It contains all the possible options.
type Input struct {
	// Version prints the version banner.
	Version bool
	// Delay between a page crawled and another.
	Delay int
	// Concurrency level.
	Concurrency int
	// Help prints the help banner.
	Help bool
	// Examples prints the examples banner.
	Examples bool
	// Plain prints only the results.
	Plain bool
	// JSON prints the output as JSON in stdout.
	JSON bool
	// HTMLout writes the output into an HTML file.
	HTMLout string
	// TXTout writes the output into an TXT file.
	TXTout string
	// Ignore ignores the URL containing at least one of the elements of this array.
	Ignore string
	// IgnoreTXT ignores the URL containing at least one of the lines of this file.
	IgnoreTXT string
	// Cache uses the .cariddi_cache folder as cache.
	Cache bool
	// Timeout set timeout for the requests. (default 10)
	Timeout int
	// Intensive crawls searching for resources matching 2nd level domain.
	Intensive bool
	// Rua uses a random browser user agent on every request.
	Rua bool
	// Proxy set a Proxy to be used (http and socks5 supported).
	Proxy string
	// Secrets hunts for secrets.
	Secrets bool
	// SecretsFile uses an external file (txt, one per line) to use custom regexes for secrets hunting.
	SecretsFile string
	// Endpoints hunts for juicy endpoints.
	Endpoints bool
	// EndpointsFile uses an external file (txt, one per line) to use custom parameters for endpoints hunting.
	EndpointsFile string
	// Extensions hunts for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
	Extensions int
	// Headers uses custom headers for each request E.g. -headers "Cookie: auth=yes;;Client: type=2".
	Headers string
	// HeadersFile reads from an external file custom headers (same format of headers flag).
	HeadersFile string
	// Errors hunts for errors in websites.
	Errors bool
	// Info hunts for useful informations in websites.
	Info bool
	// Debug prints debug information while crawling.
	Debug bool
	// UserAgent uses a custom User Agent.
	UserAgent string
	// StoreResp stores HTTP responses.
	StoreResp bool
	// StoredRespDir stores HTTP responses to the directory provided.
	StoredRespDir string
}

// ScanFlag defines all the options taken
// as input and scan them, then it returns
// an Input struct.
func ScanFlag() Input {
	versionPtr := flag.Bool("version", false, "Print the version.")
	delayPtr := flag.Int("d", 0, "Delay between a page crawled and another.")
	concurrencyPtr := flag.Int("c", DefaultConcurrency, "Concurrency level.")
	helpPtr := flag.Bool("h", false, "Print the help.")
	examplesPtr := flag.Bool("examples", false, "Print the examples.")
	plainPtr := flag.Bool("plain", false, "Print only results.")
	JSONPtr := flag.Bool("json", false, "Print the output as JSON in stdout.")
	outputHTMLPtr := flag.String("oh", "", "Write the output into an HTML file.")
	outputTXTPtr := flag.String("ot", "", "Write the output into a TXT file.")
	ignorePtr := flag.String("i", "", "Ignore the URL containing at least one of the elements of this array.")
	ignoreTXTPtr := flag.String("it", "", "Ignore the URL containing at least one of the lines of this file.")
	cachePtr := flag.Bool("cache", false, "Use the .cariddi_cache folder as cache.")
	timeoutPtr := flag.Int("t", TimeoutRequest, "Set timeout for the requests.")
	intensivePtr := flag.Bool("intensive", false, "Crawl searching for resources matching 2nd level domain.")
	ruaPtr := flag.Bool("rua", false, "Use a random browser user agent on every request.")
	proxyPtr := flag.String("proxy", "", "Set a Proxy, http and socks5 supported.")

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

	storeRespPtr := flag.Bool("sr", false, "Store HTTP responses.")

	storedRespDirPtr := flag.String("srd", "", "Stores HTTP responses to the directory provided.")

	flag.Parse()

	result := Input{
		*versionPtr,
		*delayPtr,
		*concurrencyPtr,
		*helpPtr,
		*examplesPtr,
		*plainPtr,
		*JSONPtr,
		*outputHTMLPtr,
		*outputTXTPtr,
		*ignorePtr,
		*ignoreTXTPtr,
		*cachePtr,
		*timeoutPtr,
		*intensivePtr,
		*ruaPtr,
		*proxyPtr,
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
		*storeRespPtr,
		*storedRespDirPtr,
	}

	return result
}
