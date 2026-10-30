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

import (
	"flag"
)

const (
	DefaultConcurrency      = 20
	DefaultIgnoreExtensions = "png,svg,jpg,jpeg,bmp,jfif,gif,webp,woff,woff2,ttf,tiff,tif," +
		"mp4,webm,mkv,avi,mov,flv,wmv," + // Video
		"mp3,wav,flac,ogg,m4a,aac," + // Audio
		"ico,cur,eot,otf" // More icons/fonts
	TimeoutRequest = 10
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
	// Intensive crawls searching for resources matching the root domain and its subdomains.
	Intensive bool
	// Rua uses a random browser user agent on every request.
	Rua bool
	// Proxy set a Proxy to be used (http and socks5 supported).
	// If not set, the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables are used.
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
	// HeadersFile reads from an external file custom headers (same format of headers flag, or one per line).
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
	// MaxDepth is the maximum depth to crawl: 1 follows the links
	// of the target, 2 also their links and so on (0 means no limit).
	MaxDepth int
	// IgnoreExtensions specifies which extensions must be ignored while scanning
	// (Default: png, svg, jpg, jpeg, bmp, jfif, gif, webp, woff, woff2, ttf, tiff, tif, mp4,
	// webm, mkv, avi, mov, flv, wmv, mp3, wav, flac, ogg, m4a, aac, ico, cur, eot, otf)
	IgnoreExtensions StringSlice
}

// ScanFlag defines all the options taken
// as input and scan them, then it returns
// an Input struct.
func ScanFlag() Input {
	var flags Input

	flag.BoolVar(&flags.Version, "version", false, "Print the version.")
	flag.IntVar(&flags.Delay, "d", 0, "Delay between a page crawled and another.")
	flag.IntVar(&flags.Concurrency, "c", DefaultConcurrency, "Concurrency level.")
	flag.BoolVar(&flags.Help, "h", false, "Print the help.")
	flag.BoolVar(&flags.Examples, "examples", false, "Print the examples.")
	flag.BoolVar(&flags.Plain, "plain", false, "Print only results.")
	flag.BoolVar(&flags.JSON, "json", false, "Print the output as JSON in stdout.")
	flag.StringVar(&flags.HTMLout, "oh", "", "Write the output into an HTML file.")
	flag.StringVar(&flags.TXTout, "ot", "", "Write the output into a TXT file.")
	flag.StringVar(&flags.Ignore, "i", "", "Ignore the URL containing at least one of the elements of this array.")
	flag.StringVar(&flags.IgnoreTXT, "it", "", "Ignore the URL containing at least one of the lines of this file.")
	flag.BoolVar(&flags.Cache, "cache", false, "Use the .cariddi_cache folder as cache.")
	flag.IntVar(&flags.Timeout, "t", TimeoutRequest, "Set timeout for the requests.")
	flag.BoolVar(&flags.Intensive, "intensive", false, "Crawl searching for resources matching the root domain and its subdomains.")
	flag.BoolVar(&flags.Rua, "rua", false, "Use a random browser user agent on every request.")
	flag.StringVar(&flags.Proxy, "proxy", "", "Set a Proxy, http and socks5 supported "+
		"(default: the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables).")

	flag.BoolVar(&flags.Secrets, "s", false, "Hunt for secrets.")
	flag.StringVar(&flags.SecretsFile, "sf", "", "Use an external file (txt, one per line)"+
		" to use custom regexes for secrets hunting.")

	flag.BoolVar(&flags.Endpoints, "e", false, "Hunt for juicy endpoints.")
	flag.StringVar(&flags.EndpointsFile, "ef", "", "Use an external file (txt, one per line)"+
		" to use custom parameters for endpoints hunting.")

	flag.IntVar(&flags.Extensions, "ext", 0, "Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).")

	flag.StringVar(&flags.Headers, "headers", "", "Use custom headers for each request "+
		"E.g. -headers \"Cookie: auth=yes;;Client: type=2\".")
	flag.StringVar(&flags.HeadersFile, "headersfile", "", "Read from an external file "+
		"custom headers (same format of headers flag, or one per line).")

	flag.BoolVar(&flags.Errors, "err", false, "Hunt for errors in websites.")

	flag.BoolVar(&flags.Info, "info", false, "Hunt for useful informations in websites.")

	flag.BoolVar(&flags.Debug, "debug", false, "Print debug information while crawling.")

	flag.StringVar(&flags.UserAgent, "ua", "", "Use a custom User Agent.")

	flag.BoolVar(&flags.StoreResp, "sr", false, "Store HTTP responses.")

	flag.IntVar(&flags.MaxDepth, "md", 0, "Maximum depth to crawl: 1 follows the links of the target, 2 also their links (0 = no limit).")

	flag.Var(&flags.IgnoreExtensions, "ie", "Comma-separated list of extensions to ignore while scanning")

	flag.Parse()

	// Default Extensions to filter
	if flags.IgnoreExtensions == nil {
		_ = flags.IgnoreExtensions.Set(DefaultIgnoreExtensions)
	}

	return flags
}
