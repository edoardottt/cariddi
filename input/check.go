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

	"github.com/edoardottt/cariddi/utils"
)

//CheckDataPost > TODO
func CheckDataPost(input string) (map[string]string, error) {

	// ===== TODO =======
	return map[string]string{}, nil
}

//CheckOutputFile >
func CheckOutputFile(input string) bool {
	invalid := []string{"\\", "/", "'", "\""}
	for _, elem := range invalid {
		if strings.ContainsAny(input, elem) {
			return false
		}
	}
	return true
}

//CheckFlags checks the flags inserted
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

	if flags.Extensions != 0 {
		if !(1 <= flags.Extensions && flags.Extensions <= 7) {
			fmt.Println("The extension value must go from 1 (juicy) to 7 (not juicy).")
			os.Exit(1)
		}
	}

	if flags.EndpointsFile != "" {
		if !flags.Endpoints {
			fmt.Println("You can't define an endpoint file and not the endpoint search.")
			fmt.Println("If you want to scan for custom parameters enter both -e and -ef {filename}.")
			os.Exit(1)
		}
	}

	if flags.SecretsFile != "" {
		if !flags.Secrets {
			fmt.Println("You can't define a secrets file and not the secrets search.")
			fmt.Println("If you want to scan for custom regexes enter both -s and -sf {filename}.")
			os.Exit(1)
		}
	}

	if flags.Plain && flags.Txt == "" && flags.Html == "" {
		if flags.Secrets || flags.Endpoints || flags.Extensions != 0 {
			fmt.Println("In the plain mode cariddi prints only links found on targets.")
			fmt.Println("If you want to see the results of secrets, endpoints and extensions found")
			fmt.Println("you should define a Txt or/and Html file output, or remove the plain mode.")
			fmt.Println("Examples:")
			fmt.Println("	- cat urls | cariddi -plain -s -ot {target-name}")
			fmt.Println("	- cat urls | cariddi -s")
			os.Exit(1)
		}
	}

	if flags.IgnoreTxt != "" {
		_ = utils.ReadFile(flags.IgnoreTxt)
	}

	if flags.Timeout < 0 {
		fmt.Println("The timeout value must be a positive value.")
		os.Exit(1)
	}

	if flags.Ignore != "" && flags.IgnoreTxt != "" {
		fmt.Println("You should use only one among -i and -it.")
		fmt.Println("Examples:")
		fmt.Println("	- cat urls | cariddi -i forum,blog")
		fmt.Println("	- cat urls | cariddi -it ignore.txt")
		os.Exit(1)
	}
}
