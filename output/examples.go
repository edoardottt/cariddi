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

//PrintExamples
func PrintExamples() {
	Beautify()
	fmt.Println(`
	cat urls | cariddi -version (Print the version)

	cat urls | cariddi -h (Print the help)
	
	cat urls | cariddi -e (Hunt for secrets)
	
	cat urls | cariddi -d 2 (2 seconds between a page crawled and another)
	
	cat urls | cariddi -c 200 (Set the concurrency level to 200)
	
	cat urls | cariddi -s (Hunt for juicy endpoints)
	
	cat urls | cariddi -plain (Print only useful things)
	
	cat urls | cariddi -ot target_name (Results in txt file)
	
	cat urls | cariddi -oh target_name (Results in html file)
	
	cat urls | cariddi -ext 2 (Hunt for juicy (level 2 of 7) files)
	
	cat urls | cariddi -e -ef endpoints_file (Hunt for custom endpoints)

	cat urls | cariddi -s -sf secrets_file (Hunt for custom secrets)`)
}
