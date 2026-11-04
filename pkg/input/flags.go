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
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/projectdiscovery/goflags"
)

const (
	DefaultThreads          = 20
	DefaultIgnoreExtensions = "png,svg,jpg,jpeg,bmp,jfif,gif,webp,woff,woff2,ttf,tiff,tif," +
		"mp4,webm,mkv,avi,mov,flv,wmv," + // Video
		"mp3,wav,flac,ogg,m4a,aac," + // Audio
		"ico,cur,eot,otf" // More icons/fonts
	TimeoutRequest = 10
)

// Input struct.
// It contains all the possible options.
type Input struct {
	// URLs are the URLs to crawl given with -u.
	URLs []string
	// List are the files containing the URLs to crawl, one per line.
	List []string

	// MaxDepth is the maximum depth to crawl: 1 follows the links
	// of the target, 2 also their links and so on (0 means no limit).
	MaxDepth int
	// Timeout set timeout for the requests, in seconds. (default 10)
	Timeout int
	// Headers are the custom headers for each request: each element is a file
	// containing headers or the headers themselves (see GetHeaders).
	Headers []string
	// UserAgent uses a custom User Agent.
	UserAgent string
	// Rua uses a random browser user agent on every request.
	Rua bool
	// Proxies are the proxies to use (http, https, socks5 and socks5h supported):
	// each element is a proxy or a file containing proxies, one per line (see
	// GetProxies). With more than one, each request uses the next one in turn.
	// If not set, the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables are used.
	Proxies []string
	// Cache uses the .cariddi_cache folder as cache.
	Cache bool

	// Subdomains crawls searching for resources matching the root domain and its subdomains.
	Subdomains bool
	// Ignore ignores the URLs containing at least one of these strings: each element
	// is a file (one string per line) or a comma-separated list (see GetIgnore).
	Ignore []string

	// Threads is the number of concurrent requests.
	Threads int
	// Delay between a page crawled and another, in seconds.
	Delay int

	// Secrets hunts for secrets.
	Secrets bool
	// SecretsFile is a file (one per line) with the regexes used for secrets
	// hunting instead of the default ones. It enables Secrets.
	SecretsFile string
	// Endpoints hunts for juicy endpoints.
	Endpoints bool
	// EndpointsFile is a file (one per line) with the parameters used for endpoints
	// hunting instead of the default ones. It enables Endpoints.
	EndpointsFile string
	// Extensions hunts for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
	Extensions int
	// Errors hunts for errors in websites.
	Errors bool
	// Info hunts for useful information in websites.
	Info bool
	// IgnoreExtensions specifies which extensions must be ignored while scanning
	// for secrets, errors and info (Default: png, svg, jpg, jpeg, bmp, jfif, gif,
	// webp, woff, woff2, ttf, tiff, tif, mp4, webm, mkv, avi, mov, flv, wmv, mp3,
	// wav, flac, ogg, m4a, aac, ico, cur, eot, otf)
	IgnoreExtensions []string

	// TXTout writes the output into a TXT file.
	TXTout string
	// HTMLout writes the output into an HTML file.
	HTMLout string
	// JSON prints the output as JSON lines in stdout.
	JSON bool
	// StoreResp stores HTTP responses.
	StoreResp bool
	// Silent prints only the URLs found, without the banner.
	Silent bool
	// Debug prints debug information while crawling.
	Debug bool
	// Version prints the version banner.
	Version bool
	// Examples prints the examples banner.
	Examples bool
}

// ScanFlag defines all the options taken
// as input and scan them, then it returns
// an Input struct.
// With -h or -help, it prints the help and exits (see HelpRequested).
func ScanFlag() Input {
	flags, unexpected := scanFlag(os.Args[1:])

	// The parsing stops at the first argument that is not a flag:
	// the flags after it would be silently ignored.
	if len(unexpected) > 0 {
		fmt.Printf("Unexpected argument: %s (the flags after it are not read).\n", unexpected[0])
		fmt.Println("Use -u to crawl a URL, e.g. cariddi -u https://example.com")
		os.Exit(1)
	}

	return flags
}

// scanFlag scans the flags in args, returning them and the arguments
// left after the flags. With no args, goflags reads the ones of os.Args.
func scanFlag(args []string) (Input, []string) {
	var flags Input

	flagSet := newFlagSet(&flags)

	if err := flagSet.Parse(args...); err != nil {
		fmt.Println("Can't parse the flags:", err)
		os.Exit(1)
	}

	// A custom file enables its hunt.
	flags.Secrets = flags.Secrets || flags.SecretsFile != ""
	flags.Endpoints = flags.Endpoints || flags.EndpointsFile != ""

	return flags, flagSet.CommandLine.Args()
}

// HelpRequested reports whether ScanFlag prints the help for args (the
// arguments of the command line): the flag package prints the usage on -h,
// -help and on a parse error, and the usage shows the help only if one of
// args is h or help (e.g. "-x -h"). args are parsed by a copy of the flags.
func HelpRequested(args []string) bool {
	usage := false

	commandLine := newFlagSet(&Input{}).CommandLine
	commandLine.Init(commandLine.Name(), flag.ContinueOnError)
	commandLine.SetOutput(io.Discard)
	commandLine.Usage = func() { usage = true }

	_ = commandLine.Parse(args)

	return usage && slices.ContainsFunc(args, func(arg string) bool {
		name := strings.Trim(arg, "-")

		return name == "h" || name == "help"
	})
}

// newFlagSet defines the flags of cariddi, storing their values in flags.
func newFlagSet(flags *Input) *goflags.FlagSet {
	// The options come only from the command line. Without this, the
	// first run would create a configuration file in the user's
	// configuration directory, and every run would read it.
	goflags.DisableAutoConfigMigration = true

	flagSet := goflags.NewFlagSet()
	flagSet.SetConfigFilePath(os.DevNull)

	flagSet.CreateGroup("input", "Input",
		// The URLs and the file names are not split on commas: they can contain them.
		flagSet.StringSliceVarP((*goflags.StringSlice)(&flags.URLs), "url", "u", nil,
			"URL to crawl (repeat -u to crawl more URLs).", goflags.StringSliceOptions),
		flagSet.StringSliceVarP((*goflags.StringSlice)(&flags.List), "list", "l", nil,
			"File containing the URLs to crawl, one per line (repeat -l to read more files).",
			goflags.StringSliceOptions),
	)

	flagSet.CreateGroup("configuration", "Configuration",
		flagSet.IntVarP(&flags.MaxDepth, "depth", "d", 0,
			"Maximum depth to crawl: 1 follows the links of the target, 2 also their links (0 = no limit)."),
		flagSet.IntVar(&flags.Timeout, "timeout", TimeoutRequest, "Set timeout for the requests, in seconds."),
		// The values are used as they are: GetHeaders reads the files.
		flagSet.StringSliceVarP((*goflags.StringSlice)(&flags.Headers), "headers", "H", nil,
			"Use custom headers for each request (e.g. -H \"Cookie: auth=yes;;Client: type=2\"), "+
				"or read them from a file (one per line). Repeatable.", goflags.StringSliceOptions),
		flagSet.StringVarP(&flags.UserAgent, "user-agent", "ua", "", "Use a custom User Agent."),
		flagSet.BoolVarP(&flags.Rua, "random-agent", "rua", false, "Use a random browser user agent on every request."),
		// The values are used as they are: GetProxies reads the files.
		flagSet.StringSliceVar((*goflags.StringSlice)(&flags.Proxies), "proxy", nil,
			"Set a Proxy to be used (http, https, socks5 and socks5h supported), or read them from a file (one per line). "+
				"Repeatable: each request uses the next proxy in turn. "+
				"Default: the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables.", goflags.StringSliceOptions),
		flagSet.BoolVar(&flags.Cache, "cache", false, "Use the .cariddi_cache folder as cache."),
	)

	flagSet.CreateGroup("scope", "Scope",
		flagSet.BoolVarP(&flags.Subdomains, "subdomains", "subs", false,
			"Crawl searching for resources matching the root domain and its subdomains."),
		// The values are used as they are: GetIgnore reads the files.
		flagSet.StringSliceVarP((*goflags.StringSlice)(&flags.Ignore), "ignore", "i", nil,
			"Ignore the URLs containing at least one of these comma-separated strings, "+
				"or of the lines of this file. Repeatable.", goflags.StringSliceOptions),
	)

	flagSet.CreateGroup("rate-limit", "Rate-Limit",
		flagSet.IntVarP(&flags.Threads, "threads", "t", DefaultThreads, "Number of concurrent requests."),
		flagSet.IntVarP(&flags.Delay, "delay", "rd", 0, "Delay between a page crawled and another, in seconds."),
	)

	flagSet.CreateGroup("hunt", "Hunt",
		flagSet.BoolVarP(&flags.Secrets, "secrets", "s", false, "Hunt for secrets."),
		flagSet.StringVarP(&flags.SecretsFile, "secrets-file", "sf", "",
			"Hunt for secrets with the regexes of this file (one per line) instead of the default ones."),
		flagSet.BoolVarP(&flags.Endpoints, "endpoints", "e", false, "Hunt for juicy endpoints."),
		flagSet.StringVarP(&flags.EndpointsFile, "endpoints-file", "ef", "",
			"Hunt for juicy endpoints with the parameters of this file (one per line) instead of the default ones."),
		flagSet.IntVarP(&flags.Extensions, "extensions", "ext", 0,
			"Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy)."),
		flagSet.BoolVarP(&flags.Errors, "errors", "err", false, "Hunt for errors in websites."),
		flagSet.BoolVar(&flags.Info, "info", false, "Hunt for useful information in websites."),
		// A value given with -ie replaces the default extensions (-ie "" ignores none).
		flagSet.StringSliceVarP((*goflags.StringSlice)(&flags.IgnoreExtensions), "ignore-ext", "ie",
			goflags.StringSlice{DefaultIgnoreExtensions},
			"Comma-separated list of extensions to ignore while scanning for secrets, errors and info.",
			goflags.NormalizedStringSliceOptions),
	)

	flagSet.CreateGroup("output", "Output",
		flagSet.StringVarP(&flags.TXTout, "output-txt", "ot", "", "Write the output into a TXT file."),
		flagSet.StringVarP(&flags.HTMLout, "output-html", "oh", "", "Write the output into an HTML file."),
		flagSet.BoolVarP(&flags.JSON, "jsonl", "j", false, "Print the output as JSON lines in stdout."),
		flagSet.BoolVarP(&flags.StoreResp, "store-response", "sr", false, "Store HTTP responses."),
		flagSet.BoolVar(&flags.Silent, "silent", false, "Print only the URLs found, without the banner."),
		flagSet.BoolVar(&flags.Debug, "debug", false, "Print debug information while crawling."),
		flagSet.BoolVarP(&flags.Version, "version", "v", false, "Print the version."),
		flagSet.BoolVar(&flags.Examples, "examples", false, "Print the examples."),
	)

	return flagSet
}
