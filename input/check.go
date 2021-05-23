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

package input

import (
	"fmt"
	"os"
	"strings"
)

//CheckDataPost
func CheckDataPost(input string) (map[string]string, error) {

	// ===== TODO =======
	return map[string]string{}, nil
}

//CheckOutputFile
func CheckOutputFile(input string) bool {
	invalid := []string{"\\", "/", "'", "\""}
	for _, elem := range invalid {
		if strings.ContainsAny(input, elem) {
			return false
		}
	}
	return true
}

//CheckFlags
func CheckFlags(flags Input) {
	if flags.Txt != "" {
		if !CheckOutputFile(flags.Txt) {
			fmt.Println("The output file must avoid weird symbols. Try to use - , _ , . instead.")
			os.Exit(1)
		}
	}

	if flags.Html != "" {
		if !CheckOutputFile(flags.Html) {
			fmt.Println("The output file must avoid weird symbols. Try to use - , _ , . instead.")
			os.Exit(1)
		}
	}
}
