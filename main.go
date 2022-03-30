/*
==========
Cariddi v1.1.6
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

package main

import (
	"os"

	"github.com/edoardottt/cariddi/crawler"
	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/edoardottt/cariddi/utils"
)

//main function >
func main() {

	// Scan flags.
	flags := input.ScanFlag()

	//Print version and exit.
	if flags.Version {
		output.Beautify()
		os.Exit(0)
	}

	//Print help and exit.
	if flags.Help {
		output.PrintHelp()
		os.Exit(0)
	}

	//Print examples and exit.
	if flags.Examples {
		output.PrintExamples()
		os.Exit(0)
	}

	//If it's possible print the cariddi banner.
	if !flags.Plain {
		output.Beautify()
	}

	//Read the targets from standard input.
	targets := input.ScanTargets()

	//Check if there are errors in the flags definition.
	input.CheckFlags(flags)

	//If it is needed, read custom endpoints definition
	//from the specified file.
	var endpointsFileSlice []string
	if flags.EndpointsFile != "" {
		endpointsFileSlice = utils.ReadFile(flags.EndpointsFile)
	}

	//If it is needed, read custom secrets definition
	//from the specified file.
	var secretsFileSlice []string
	if flags.SecretsFile != "" {
		secretsFileSlice = utils.ReadFile(flags.SecretsFile)
	}

	var finalResults []string
	var finalSecret []scanner.SecretMatched
	var finalEndpoints []scanner.EndpointMatched
	var finalExtensions []scanner.FileTypeMatched
	var finalErrors []scanner.ErrorMatched
	var finalInfos []scanner.InfoMatched

	//Create output files if needed (txt / html).
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

	//Read headers if needed
	var headers map[string]string
	if flags.HeadersFile != "" || flags.Headers != "" {
		var headersInput string
		if flags.HeadersFile != "" {
			headersInput = string(utils.ReadEntireFile(flags.HeadersFile))
		} else {
			headersInput = flags.Headers
		}
		headers = input.GetHeaders(headersInput)
	}

	//For each target generate a crawler and collect all the results.
	for _, inp := range targets {

		results, secrets, endpoints, extensions, errors, infos := crawler.Crawler(inp, ResultTxt, ResultHtml, flags.Delay,
			flags.Concurrency, flags.Ignore, flags.IgnoreTxt, flags.Cache, flags.Timeout, flags.Intensive,
			flags.Rua, flags.Proxy, flags.Secrets, secretsFileSlice, flags.Plain, flags.Endpoints, endpointsFileSlice,
			flags.Extensions, headers, flags.Errors, flags.Info)

		finalResults = append(finalResults, results...)
		finalSecret = append(finalSecret, secrets...)
		finalEndpoints = append(finalEndpoints, endpoints...)
		finalExtensions = append(finalExtensions, extensions...)
		finalErrors = append(finalErrors, errors...)
		finalInfos = append(finalInfos, infos...)
	}

	//Remove duplicates from all the results.
	finalResults = utils.RemoveDuplicateValues(finalResults)
	finalSecret = scanner.RemoveDuplicateSecrets(finalSecret)
	finalEndpoints = scanner.RemovDuplicateEndpoints(finalEndpoints)
	finalExtensions = scanner.RemoveDuplicateExtensions(finalExtensions)
	finalErrors = scanner.RemoveDuplicateErrors(finalErrors)
	finalInfos = scanner.RemoveDuplicateInfos(finalInfos)

	// IF TXT OUTPUT >
	if flags.Txt != "" {
		output.TxtOutput(flags, finalResults, finalSecret, finalEndpoints,
			finalExtensions, finalErrors)
	}

	// IF HTML OUTPUT >
	if flags.Html != "" {
		output.HtmlOutput(flags, ResultHtml, finalResults, finalSecret,
			finalEndpoints, finalExtensions, finalErrors)
	}

	//If needed print secrets.
	if !flags.Plain && len(finalSecret) != 0 {
		for _, elem := range finalSecret {
			output.EncapsulateCustomGreen(elem.Secret.Name, elem.Match+" in "+elem.Url)
		}
	}

	//If needed print endpoints.
	if !flags.Plain && len(finalEndpoints) != 0 {
		for _, elem := range finalEndpoints {
			for _, parameter := range elem.Parameters {
				finalString := ""
				finalString += parameter.Parameter
				if len(parameter.Attacks) != 0 {
					finalString += " -"
					for _, attack := range parameter.Attacks {
						finalString += " " + attack
					}
				}
				output.EncapsulateCustomGreen(finalString, " in "+elem.Url)
			}
		}
	}

	//If needed print extensions.
	if !flags.Plain && len(finalExtensions) != 0 {
		for _, elem := range finalExtensions {
			output.EncapsulateCustomGreen(elem.Filetype.Extension, elem.Url+" matched!")
		}
	}

	//If needed print errors.
	if !flags.Plain && len(finalErrors) != 0 {
		for _, elem := range finalErrors {
			output.EncapsulateCustomGreen(elem.Error.ErrorName, elem.Match+" in "+elem.Url)
		}
	}
}
