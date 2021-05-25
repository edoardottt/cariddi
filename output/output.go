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

import (
	"fmt"
	"os"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/edoardottt/cariddi/utils"
)

//PrintOutput
func PrintSimpleOutput(out []string) {
	for _, elem := range out {
		fmt.Println(elem)
	}
}

//TxtOutput it's the wrapper around all the txt things.
//Actually it manages everything related to TXT output.
func TxtOutput(flags input.Input, finalSecret []scanner.SecretMatched, finalEndpoints []scanner.EndpointMatched, finalExtensions []scanner.FileTypeMatched) {

	exists, err := utils.ElementExists("output-cariddi")
	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}
	if !exists {
		utils.CreateOutputFolder()
	}

	// if secrets flag enabled save also secrets
	if flags.Secrets {
		SecretFilename := utils.CreateOutputFile(flags.Txt, "secrets", "txt")
		for _, elem := range finalSecret {
			AppendOutputToTxt(elem.Secret.Name+" Found in "+elem.Url+" "+elem.Secret.Regex, SecretFilename)
		}
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints {
		EndpointFilename := utils.CreateOutputFile(flags.Txt, "endpoints", "txt")
		for _, elem := range finalEndpoints {
			finalString := ""
			for _, parameter := range elem.Parameters {
				finalString += parameter
			}
			AppendOutputToTxt(finalString+" Found in "+elem.Url, EndpointFilename)
		}
	}

	// if extensions flag enabled save also secrets
	if 1 <= flags.Extensions && flags.Extensions <= 7 {
		ExtensionsFilename := utils.CreateOutputFile(flags.Txt, "extensions", "txt")
		for _, elem := range finalExtensions {
			AppendOutputToTxt(elem.Filetype.Extension+" Found in "+elem.Url, ExtensionsFilename)
		}
	}

}

//HtmlOutput it's the wrapper around all the html things.
//Actually it manages everything related to HTML output.
func HtmlOutput(flags input.Input, ResultFilename string, finalSecret []scanner.SecretMatched,
	finalEndpoints []scanner.EndpointMatched, finalExtensions []scanner.FileTypeMatched) {
	exists, err := utils.ElementExists("output-cariddi")

	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}

	if !exists {
		utils.CreateOutputFolder()
	}

	// if secrets flag enabled save also secrets
	if flags.Secrets {
		HeaderHTML("Secrets found", ResultFilename)
		for _, elem := range finalSecret {
			AppendOutputToHTML(elem.Secret.Name+" Found in "+elem.Url+" "+elem.Secret.Regex, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints {
		HeaderHTML("Endpoints found", ResultFilename)
		for _, elem := range finalEndpoints {
			finalString := ""
			for _, parameter := range elem.Parameters {
				finalString += parameter
			}
			AppendOutputToHTML(finalString+" Found in "+elem.Url, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	// if extensions flag enabled save also extensions
	if 1 <= flags.Extensions && flags.Extensions <= 7 {
		HeaderHTML("Extensions found", ResultFilename)
		for _, elem := range finalExtensions {
			AppendOutputToHTML(elem.Filetype.Extension+" Found in "+elem.Url, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	BannerFooterHTML(ResultFilename)
}
