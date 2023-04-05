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

package output

import "fmt"

// PrintHelp prints the help.
func PrintHelp() {
	Banner()
	fmt.Println(`Usage of cariddi:
	-c int
		Concurrency level. (default 20)
	-cache
		Use the .cariddi_cache folder as cache.
	-d int
		Delay between a page crawled and another.
	-debug
		Print debug information while crawling.
	-e	Hunt for juicy endpoints.
	-ef string
		Use an external file (txt, one per line) to use custom parameters for endpoints hunting.
	-err
		Hunt for errors in websites.
	-examples
		Print the examples.
	-ext int
		Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
	-h	Print the help.
	-headers string
		Use custom headers for each request E.g. -headers "Cookie: auth=yes;;Client: type=2".
  	-headersfile string
	  	Read from an external file custom headers (same format of headers flag).
	-json
		Print the output as JSON in stdout.
	-i string
		Ignore the URL containing at least one of the elements of this array.
	-info
		Hunt for useful informations in websites.
	-intensive
		Crawl searching for resources matching 2nd level domain.
	-it string
		Ignore the URL containing at least one of the lines of this file.
	-oh string
		Write the output into an HTML file.
	-ot string
		Write the output into a TXT file.
	-plain
		Print only the results.
	-proxy string
		Set a Proxy to be used (http and socks5 supported).
	-rua
		Use a random browser user agent on every request.
	-s	Hunt for secrets.
	-sf string
		Use an external file (txt, one per line) to use custom regexes for secrets hunting.
	-t int
		Set timeout for the requests. (default 10)
	-ua
		Use a custom User Agent.
	-version
		Print the version.`)
}
