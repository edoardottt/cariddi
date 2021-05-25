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
	"github.com/edoardottt/cariddi/utils"
)

//main
func main() {

	targets := input.ScanTargets()
	flags := input.ScanFlag()

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

	var endpointsFileSlice []string
	if flags.EndpointsFile != "" {
		endpointsFileSlice = utils.ReadFile(flags.EndpointsFile)
	}

	var secretsFileSlice []string
	if flags.SecretsFile != "" {
		secretsFileSlice = utils.ReadFile(flags.SecretsFile)
	}

	var finalSecret []scanner.SecretMatched
	var finalEndpoints []scanner.EndpointMatched
	var finalExtensions []scanner.FileTypeMatched

	// output files
	var ResultTxt = ""
	if flags.Txt != "" {
		ResultTxt = utils.CreateOutputFile(flags.Txt, "results", "txt")
	}
	var ResultHtml = ""
	if flags.Html != "" {
		ResultHtml = utils.CreateOutputFile(flags.Html, "", "html")
		output.BannerHTML(ResultHtml)
		output.HeaderHTML("Results", ResultHtml)
	}

	for _, inp := range targets {

		secrets, endpoints, extensions := crawler.Crawler(inp, ResultTxt, ResultHtml, flags.Delay, flags.Concurrency, flags.Secrets,
			secretsFileSlice, flags.Plain, flags.Endpoints, endpointsFileSlice, flags.Extensions)

		finalSecret = append(finalSecret, secrets...)
		finalEndpoints = append(finalEndpoints, endpoints...)
		finalExtensions = append(finalExtensions, extensions...)
	}

	finalSecret = scanner.RemoveDuplicateSecrets(finalSecret)
	finalEndpoints = scanner.RemovDuplicateEndpoints(finalEndpoints)
	finalExtensions = scanner.RemoveDuplicateExtensions(finalExtensions)

	// IF TXT OUTPUT
	if flags.Txt != "" {
		output.TxtOutput(flags, finalSecret, finalEndpoints, finalExtensions)
	}

	// IF HTML OUTPUT
	if flags.Html != "" {
		output.HtmlOutput(flags, ResultHtml, finalSecret, finalEndpoints, finalExtensions)
	}

	// if needed print secrets
	if !flags.Plain && len(finalSecret) != 0 {
		for _, elem := range finalSecret {
			output.EncapsulateCustomGreen(elem.Secret.Name, "Found in "+elem.Url+" "+elem.Secret.Regex+" matched!")
		}
	}

	// if needed print endpoints
	if !flags.Plain && len(finalEndpoints) != 0 {
		for _, elem := range finalEndpoints {
			finalString := ""
			for _, parameter := range elem.Parameters {
				finalString += parameter
			}
			output.EncapsulateCustomGreen(finalString, "Found in "+elem.Url+" matched!")
		}
	}

	// if needed print extensions
	if !flags.Plain && len(finalExtensions) != 0 {
		for _, elem := range finalExtensions {
			output.EncapsulateCustomGreen(elem.Filetype.Extension, elem.Url+" matched!")
		}
	}
}
