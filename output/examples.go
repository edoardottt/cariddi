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

//PrintExamples prints some examples
func PrintExamples() {
	Beautify()
	fmt.Println(`
	cariddi -version (Print the version)

	cariddi -h (Print the help)

	cariddi -examples (Print the examples)
	
	cat urls | cariddi -e (Hunt for secrets)
	
	cat urls | cariddi -d 2 (2 seconds between a page crawled and another)
	
	cat urls | cariddi -c 200 (Set the concurrency level to 200)
	
	cat urls | cariddi -s (Hunt for juicy endpoints)
	
	cat urls | cariddi -plain (Print only useful things)
	
	cat urls | cariddi -ot target_name (Results in txt file)
	
	cat urls | cariddi -oh target_name (Results in html file)
	
	cat urls | cariddi -ext 2 (Hunt for juicy (level 2 out of 7) files)
	
	cat urls | cariddi -e -ef endpoints_file (Hunt for custom endpoints)

	cat urls | cariddi -s -sf secrets_file (Hunt for custom secrets)
	
	cat urls | cariddi -i forum,blog,community,open (Ignore urls containing these words)
	
	cat urls | cariddi -it ignore_file (Ignore urls containing at least one line in the input file.)
	
	cat urls | cariddi -cache (Use the .cariddi_cache folder as cache)

	cat urls | cariddi -t 5 (Set the timeout for the requests)

	cat urls | cariddi -intensive (Crawl searching for resources matching 2nd level domain)

	cat urls | cariddi -rua (Use a random browser user agent on every request)

	cat urls | cariddi -proxy http://127.0.0.1:8080 (Set a Proxy to be used (http and socks5 supported))
	
	cat urls | cariddi -headers "Cookie: auth=admin;type=2;; X-Custom: customHeader"
	
	cat urls | cariddi -headersfile headers.txt

	cat urls | cariddi -err
	
	cat urls | cariddi -info
	
	cat urls | cariddi -debug`)
}
