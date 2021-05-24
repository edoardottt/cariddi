/*
==========
Cariddi v0.dev
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

package main

import (
	"os"

	"github.com/edoardottt/cariddi/crawler"
	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
	"github.com/edoardottt/cariddi/scanner"
)

//main
func main() {

	targets := input.ScanTargets()
	flags := input.ScanFlag()

	/*
		fmt.Println("FLAGS:")
		fmt.Println(flags)
		fmt.Println("--------------")
	*/

	if flags.Version {
		output.Beautify()
		os.Exit(0)
	}

	if flags.Help {
		output.PrintHelp()
		os.Exit(0)
	}

	if flags.Examples {
		output.PrintExamples()
		os.Exit(0)
	}

	if !flags.Plain {
		output.Beautify()
	}

	// ----------- TODO: check ALL input -------------------
	input.CheckFlags(flags)

	var finalResult []string
	var finalSecret []scanner.SecretMatched
	var finalEndpoints []scanner.EndpointMatched
	for _, inp := range targets {
		result, secrets, endpoints := crawler.Crawler(inp, flags.Delay, flags.Concurrency, flags.Secrets, flags.SecretsFile, flags.Plain, flags.Endpoints, flags.EndpointsFile)
		finalResult = append(finalResult, result...)
		finalSecret = append(finalSecret, secrets...)
		finalEndpoints = append(finalEndpoints, endpoints...)
	}

	// IF TXT OUTPUT
	if flags.Txt != "" {
		output.TxtOutput(flags, finalResult, finalSecret)
	}

	// IF HTML OUTPUT
	if flags.Html != "" {
		output.HtmlOutput(flags, finalResult, finalSecret)
	}

	// if needed print urls
	if !flags.Plain {
		output.PrintSimpleOutput(finalResult)
	}

	// if needed print secrets
	if !flags.Plain && len(finalSecret) != 0 {
		for _, elem := range finalSecret {
			output.EncapsulateCustomGreen(elem.Secret.Name, "Found in "+elem.Url+" "+elem.Secret.Regex+" matched!")
		}
	}

	// if needed print endpoints
	if !flags.Plain && len(finalEndpoints) != 0 {
		finalEndpoints = scanner.RemovDuplicateEndpoints(finalEndpoints)
		for _, elem := range finalEndpoints {
			finalString := ""
			for _, parameter := range elem.Parameters {
				finalString += parameter
			}
			output.EncapsulateCustomGreen(finalString, " Found in "+elem.Url+" matched!")
		}
	}
}
