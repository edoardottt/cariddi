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

import (
	"fmt"
	"os"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/edoardottt/cariddi/utils"
)

//PrintSimpleOutput prints line by line
func PrintSimpleOutput(out []string) {
	for _, elem := range out {
		fmt.Println(elem)
	}
}

//TxtOutput it's the wrapper around all the txt things.
//Actually it manages everything related to TXT output.
func TxtOutput(flags input.Input, finalResults []string, finalSecret []scanner.SecretMatched,
	finalEndpoints []scanner.EndpointMatched, finalExtensions []scanner.FileTypeMatched,
	finalErrors []scanner.ErrorMatched) {

	exists, err := utils.ElementExists("output-cariddi")
	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}
	if !exists {
		utils.CreateOutputFolder()
	}

	ResultFilename := utils.CreateOutputFile(flags.Txt, "results", "txt")
	for _, elem := range finalResults {
		AppendOutputToTxt(elem, ResultFilename)
	}

	// if secrets flag enabled save also secrets
	if flags.Secrets {
		SecretFilename := utils.CreateOutputFile(flags.Txt, "secrets", "txt")
		for _, elem := range finalSecret {
			AppendOutputToTxt(elem.Secret.Name+" - "+elem.Match+" in "+elem.Url, SecretFilename)
		}
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints {
		EndpointFilename := utils.CreateOutputFile(flags.Txt, "endpoints", "txt")
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
				AppendOutputToTxt(finalString+" in "+elem.Url, EndpointFilename)
			}
		}
	}

	// if extensions flag enabled save also secrets
	if 1 <= flags.Extensions && flags.Extensions <= 7 {
		ExtensionsFilename := utils.CreateOutputFile(flags.Txt, "extensions", "txt")
		for _, elem := range finalExtensions {
			AppendOutputToTxt(elem.Filetype.Extension+" in "+elem.Url, ExtensionsFilename)
		}
	}

	// if errors flag enabled save also errors
	if flags.Errors {
		ErrorsFilename := utils.CreateOutputFile(flags.Txt, "errors", "txt")
		for _, elem := range finalErrors {
			AppendOutputToTxt(elem.Error.ErrorName+" - "+elem.Match+" in "+elem.Url, ErrorsFilename)
		}
	}

}

//HtmlOutput it's the wrapper around all the html things.
//Actually it manages everything related to HTML output.
func HtmlOutput(flags input.Input, ResultFilename string, finalResults []string, finalSecret []scanner.SecretMatched,
	finalEndpoints []scanner.EndpointMatched, finalExtensions []scanner.FileTypeMatched,
	finalErrors []scanner.ErrorMatched) {
	exists, err := utils.ElementExists("output-cariddi")

	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}

	if !exists {
		utils.CreateOutputFolder()
	}

	HeaderHTML("Results found", ResultFilename)
	for _, elem := range finalResults {
		AppendOutputToHTML(elem, "", ResultFilename, true)
	}
	FooterHTML(ResultFilename)

	// if secrets flag enabled save also secrets
	if flags.Secrets {
		HeaderHTML("Secrets found", ResultFilename)
		for _, elem := range finalSecret {
			AppendOutputToHTML(elem.Secret.Name+" - "+elem.Match+" in "+elem.Url, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints {
		HeaderHTML("Endpoints found", ResultFilename)
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
				AppendOutputToHTML(finalString+" in "+elem.Url, "", ResultFilename, false)
			}
		}
		FooterHTML(ResultFilename)
	}

	// if extensions flag enabled save also extensions
	if 1 <= flags.Extensions && flags.Extensions <= 7 {
		HeaderHTML("Extensions found", ResultFilename)
		for _, elem := range finalExtensions {
			AppendOutputToHTML(elem.Filetype.Extension+" in "+elem.Url, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	// if errors flag enabled save also errors
	if flags.Errors {
		HeaderHTML("Errors found", ResultFilename)
		for _, elem := range finalErrors {
			AppendOutputToHTML(elem.Error.ErrorName+" - "+elem.Match+" in "+elem.Url, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	BannerFooterHTML(ResultFilename)
}
