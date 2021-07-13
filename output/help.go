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
*/

package output

import "fmt"

//PrintHelp prints the help.
func PrintHelp() {
	Beautify()
	fmt.Println(`Usage of cariddi:
	-c int
		Concurrency level. (default 20)
	-cache
		Use the .cariddi_cache folder as cache.
	-d int
		Delay between a page crawled and another.
	-e	Hunt for juicy endpoints.
	-ef string
		Use an external file (txt, one per line) to use custom parameters for endpoints hunting.
	-examples
		Print the examples.
	-ext int
		Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
	-h	Print the help.
	-i string
		Ignore the URL containing at least one of the elements of this array.
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
	-s	Hunt for secrets.
	-sf string
		Use an external file (txt, one per line) to use custom regexes for secrets hunting.
	-t int
		Set timeout for the requests. (default 10)
	-version
		Print the version.`)
}
