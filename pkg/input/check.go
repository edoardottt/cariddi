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
	"bufio"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/http/httpproxy"
)

// CheckOutputFile checks if the string provided as input
// is formatted in a correct way.
func CheckOutputFile(input string) bool {
	invalid := []string{"\\", "/", "'", "\""}
	for _, elem := range invalid {
		if strings.ContainsAny(input, elem) {
			return false
		}
	}

	return true
}

// CheckProxy checks if the proxy provided as input is a URL with a host
// and a scheme supported by the HTTP client (http, https, socks5 or socks5h).
// E.g. 127.0.0.1:8080 has no scheme and localhost:8080 has "localhost" as
// scheme: every request would fail.
func CheckProxy(input string) bool {
	u, err := url.Parse(input)
	if err != nil || u.Hostname() == "" {
		return false
	}

	return slices.Contains([]string{"http", "https", "socks5", "socks5h"}, u.Scheme)
}

// checkSecretsFile checks the custom regexes of the secrets file, one per
// line (blank lines are ignored): it must contain at least one regex and all
// of them must be valid, or the scan would silently use fewer regexes (or
// none). If not, it prints every problem found and exits.
func checkSecretsFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println("Can't open the secrets file:", err)
		os.Exit(1)
	}

	var (
		invalid []string
		regexes int
		line    int
	)

	sc := bufio.NewScanner(file)
	for sc.Scan() {
		line++

		if strings.TrimSpace(sc.Text()) == "" {
			continue
		}

		regexes++

		if _, err := regexp.Compile(sc.Text()); err != nil {
			invalid = append(invalid, fmt.Sprintf("line %d: %s", line, err))
		}
	}

	file.Close()

	if err := sc.Err(); err != nil {
		if !errors.Is(err, bufio.ErrTooLong) {
			fmt.Println("Can't read the secrets file:", err)
			os.Exit(1)
		}

		invalid = append(invalid, fmt.Sprintf("line %d: too long (more than %d bytes)", line+1, bufio.MaxScanTokenSize))
	}

	if len(invalid) > 0 {
		fmt.Printf("The secrets file %s contains invalid regexes:\n", filename)

		for _, problem := range invalid {
			fmt.Println("	- " + problem)
		}

		os.Exit(1)
	}

	if regexes == 0 {
		fmt.Printf("The secrets file %s is empty: it must contain the regexes to hunt for, one per line.\n", filename)
		os.Exit(1)
	}
}

// checkProxyEnvironment checks the proxy environment variables, used when
// -proxy is not set: an invalid proxy would make every request fail. As the
// HTTP client does, a proxy without a scheme is an http one (e.g. 127.0.0.1:8080).
func checkProxyEnvironment() {
	env := httpproxy.FromEnvironment()

	for _, variable := range []struct{ name, value string }{
		{"HTTP_PROXY (or http_proxy)", env.HTTPProxy},
		{"HTTPS_PROXY (or https_proxy)", env.HTTPSProxy},
	} {
		proxy := variable.value
		if proxy != "" && !strings.Contains(proxy, "://") {
			proxy = "http://" + proxy
		}

		if proxy != "" && !CheckProxy(proxy) {
			fmt.Printf("The %s environment variable is not a valid proxy: %s\n", variable.name, variable.value)
			fmt.Println("It must be a URL with scheme http, https, socks5 or socks5h (e.g. http://127.0.0.1:8080).")
			os.Exit(1)
		}
	}
}

// CheckFlags checks the flags taken as input.
func CheckFlags(flags Input) {
	if flags.TXTout != "" {
		if !CheckOutputFile(flags.TXTout) {
			fmt.Println("The output file must avoid weird symbols. Try to use - , _ , . instead.")
			os.Exit(1)
		}
	}

	if flags.HTMLout != "" {
		if !CheckOutputFile(flags.HTMLout) {
			fmt.Println("The output file must avoid weird symbols. Try to use - , _ , . instead.")
			os.Exit(1)
		}
	}

	if flags.Extensions != 0 {
		if flags.Extensions < 1 || flags.Extensions > 7 {
			fmt.Println("The extension value must go from 1 (juicy) to 7 (not juicy).")
			os.Exit(1)
		}
	}

	if flags.SecretsFile != "" {
		checkSecretsFile(flags.SecretsFile)
	}

	if flags.Silent && !flags.JSON && flags.TXTout == "" && flags.HTMLout == "" {
		if flags.Secrets || flags.Endpoints || flags.Extensions != 0 || flags.Errors || flags.Info {
			fmt.Println("In silent mode cariddi prints only the URLs found on targets.")
			fmt.Println("If you want to see the secrets, endpoints, extensions, errors and info found")
			fmt.Println("you should define a Txt or/and Html file output, use -jsonl, or remove the silent mode.")
			fmt.Println("Examples:")
			fmt.Println("	- cat urls | cariddi -silent -s -ot {target-name}")
			fmt.Println("	- cat urls | cariddi -s")
			os.Exit(1)
		}
	}

	if flags.Silent && flags.Debug {
		fmt.Println("You cannot use both silent and debug mode.")
		os.Exit(1)
	}

	if flags.Timeout < 0 {
		fmt.Println("The timeout value must be a positive value.")
		os.Exit(1)
	}

	if flags.UserAgent != "" && flags.Rua {
		fmt.Println("You should use only one among -ua and -rua.")
		fmt.Println("Examples:")
		fmt.Println("	- cat urls | cariddi -ua \"Custom User Agent\"")
		fmt.Println("	- cat urls | cariddi -rua")
		os.Exit(1)
	}

	if flags.MaxDepth < 0 {
		fmt.Println("The depth cannot be less than 0.")
		os.Exit(1)
	}
}
