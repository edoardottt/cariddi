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

package main

import (
	"fmt"
	"os"
	"strings"

	fileUtils "github.com/edoardottt/cariddi/internal/file"
	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
	"github.com/edoardottt/cariddi/pkg/crawler"
	"github.com/edoardottt/cariddi/pkg/input"
	"github.com/edoardottt/cariddi/pkg/output"
	"github.com/edoardottt/cariddi/pkg/scanner"
)

// main function.
func main() {
	// Scan flags.
	flags := input.ScanFlag()

	// Print version and exit.
	if flags.Version {
		output.Banner()
		os.Exit(0)
	}

	// Print help and exit.
	if flags.Help {
		output.PrintHelp()
		os.Exit(0)
	}

	// Print examples and exit.
	if flags.Examples {
		output.PrintExamples()
		os.Exit(0)
	}

	// If it's possible print the cariddi banner.
	if !flags.Plain {
		output.Banner()
	}

	// Setup the config according to the flags that were
	// passed via the CLI
	config := &crawler.Scan{
		Delay:                 flags.Delay,
		Concurrency:           flags.Concurrency,
		Ignore:                flags.Ignore,
		IgnoreTxt:             flags.IgnoreTXT,
		Cache:                 flags.Cache,
		JSON:                  flags.JSON,
		Timeout:               flags.Timeout,
		Intensive:             flags.Intensive,
		Rua:                   flags.Rua,
		Proxy:                 flags.Proxy,
		SecretsFlag:           flags.Secrets,
		SecretExtensionFilter: strings.Split(flags.SecretExtensionFilter, ","),
		Plain:                 flags.Plain,
		EndpointsFlag:         flags.Endpoints,
		FileType:              flags.Extensions,
		ErrorsFlag:            flags.Errors,
		InfoFlag:              flags.Info,
		Debug:                 flags.Debug,
		UserAgent:             flags.UserAgent,
		StoreResp:             flags.StoreResp,
	}

	// Read the targets from standard input.
	targets := input.ScanTargets()

	// Check if there are errors in the flags definition.
	input.CheckFlags(flags)

	// If it is needed, read custom endpoints definition
	// from the specified file.
	if flags.EndpointsFile != "" {
		config.EndpointsSlice = fileUtils.ReadFile(flags.EndpointsFile)
	}

	// If it is needed, read custom secrets definition
	// from the specified file.
	if flags.SecretsFile != "" {
		config.SecretsSlice = fileUtils.ReadFile(flags.SecretsFile)
	}

	finalResults := []string{}
	finalSecret := []scanner.SecretMatched{}
	finalEndpoints := []scanner.EndpointMatched{}
	finalExtensions := []scanner.FileTypeMatched{}
	finalErrors := []scanner.ErrorMatched{}
	finalInfos := []scanner.InfoMatched{}

	// Create output files if needed (txt / html).
	config.Txt = ""
	if flags.TXTout != "" {
		config.Txt = fileUtils.CreateOutputFile(flags.TXTout, "results", "txt")
	}

	var ResultHTML = ""
	if flags.HTMLout != "" {
		ResultHTML = fileUtils.CreateOutputFile(flags.HTMLout, "", "html")
		output.BannerHTML(ResultHTML)
		output.HeaderHTML("Results", ResultHTML)
	}

	if config.StoreResp {
		fileUtils.CreateIndexOutputFile("index.responses.txt")
	}

	// Read headers if needed
	if flags.HeadersFile != "" || flags.Headers != "" {
		var headersInput string
		if flags.HeadersFile != "" {
			headersInput = string(fileUtils.ReadEntireFile(flags.HeadersFile))
		} else {
			headersInput = flags.Headers
		}

		config.Headers = input.GetHeaders(headersInput)
	}

	// For each target generate a crawler and collect all the results.
	for _, target := range targets {
		config.Target = target
		results := crawler.New(config)
		finalResults = append(finalResults, results.URLs...)
		finalSecret = append(finalSecret, results.Secrets...)
		finalEndpoints = append(finalEndpoints, results.Endpoints...)
		finalExtensions = append(finalExtensions, results.Extensions...)
		finalErrors = append(finalErrors, results.Errors...)
		finalInfos = append(finalInfos, results.Infos...)
	}

	// Remove duplicates from all the results.
	finalResults = sliceUtils.RemoveDuplicateValues(finalResults)
	finalSecret = scanner.RemoveDuplicateSecrets(finalSecret)
	finalEndpoints = scanner.RemovDuplicateEndpoints(finalEndpoints)
	finalExtensions = scanner.RemoveDuplicateExtensions(finalExtensions)
	finalErrors = scanner.RemoveDuplicateErrors(finalErrors)
	finalInfos = scanner.RemoveDuplicateInfos(finalInfos)

	// IF TXT OUTPUT >
	if flags.TXTout != "" {
		output.TxtOutput(flags, finalResults, finalSecret, finalEndpoints,
			finalExtensions, finalErrors, finalInfos)
	}

	// IF HTML OUTPUT >
	if flags.HTMLout != "" {
		output.HTMLOutput(flags, ResultHTML, finalResults, finalSecret,
			finalEndpoints, finalExtensions, finalErrors, finalInfos)
	}

	// If needed print secrets.
	if !flags.JSON && !flags.Plain && len(finalSecret) != 0 {
		for _, elem := range finalSecret {
			output.EncapsulateCustomGreen(elem.Secret.Name, fmt.Sprintf("%s in %s", elem.Match, elem.URL))
		}
	}

	// If needed print endpoints.
	if !flags.JSON && !flags.Plain && len(finalEndpoints) != 0 {
		for _, elem := range finalEndpoints {
			for _, parameter := range elem.Parameters {
				finalString := "" + parameter.Parameter
				if len(parameter.Attacks) != 0 {
					finalString += " -"
					for _, attack := range parameter.Attacks {
						finalString += " " + attack
					}
				}

				output.EncapsulateCustomGreen(finalString, fmt.Sprintf(" in %s", elem.URL))
			}
		}
	}

	// If needed print extensions.
	if !flags.JSON && !flags.Plain && len(finalExtensions) != 0 {
		for _, elem := range finalExtensions {
			output.EncapsulateCustomGreen(elem.Filetype.Extension, fmt.Sprintf("%s matched!", elem.URL))
		}
	}

	// If needed print errors.
	if !flags.JSON && !flags.Plain && len(finalErrors) != 0 {
		for _, elem := range finalErrors {
			output.EncapsulateCustomGreen(elem.Error.ErrorName, fmt.Sprintf("%s in %s", elem.Match, elem.URL))
		}
	}

	// If needed print infos.
	if !flags.JSON && !flags.Plain && len(finalInfos) != 0 {
		for _, elem := range finalInfos {
			output.EncapsulateCustomGreen(elem.Info.Name, fmt.Sprintf("%s in %s", elem.Match, elem.URL))
		}
	}
}
