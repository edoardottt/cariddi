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

	@Author:      edoardottt, https://www.edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package output

import (
	"fmt"
	"html"
	"os"
	"strings"

	fileUtils "github.com/edoardottt/cariddi/internal/file"
	"github.com/edoardottt/cariddi/pkg/input"
	"github.com/edoardottt/cariddi/pkg/scanner"
)

const (
	CariddiOutputFolder = "output-cariddi"
	ResultFile          = "results"
	SecretsFile         = "secrets"
	EndpointsFile       = "endpoints"
	ExtensionsFile      = "extensions"
	ErrorsFile          = "errors"
	InfosFile           = "info"
)

func hasValidExtensionFlag(flag int) bool {
	return flag >= 1 && flag <= 7
}

// PrintSimpleOutput prints line by line.
func PrintSimpleOutput(out []string) {
	for _, elem := range out {
		fmt.Println(elem)
	}
}

// TxtOutput it's the wrapper around all the txt things.
// Actually it manages everything related to TXT output.
func TxtOutput(flags input.Input, finalResults []string, finalSecret []scanner.SecretMatched,
	finalEndpoints []scanner.EndpointMatched, finalExtensions []scanner.FileTypeMatched,
	finalErrors []scanner.ErrorMatched, finalInfos []scanner.InfoMatched) {
	exists, err := fileUtils.ElementExists(CariddiOutputFolder)
	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}

	if !exists {
		fileUtils.CreateOutputFolder()
	}

	ResultFilename := fileUtils.CreateOutputFile(flags.TXTout, ResultFile, "txt")
	WriteAllTxt(finalResults, ResultFilename)

	// if secrets flag enabled save also secrets
	if flags.Secrets && len(finalSecret) > 0 {
		SecretFilename := fileUtils.CreateOutputFile(flags.TXTout, SecretsFile, "txt")
		for _, elem := range finalSecret {
			AppendOutputToTxt(fmt.Sprintf("%s - %s in %s", elem.Secret.Name, elem.Match, elem.URL), SecretFilename)
		}
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints && len(finalEndpoints) > 0 {
		EndpointFilename := fileUtils.CreateOutputFile(flags.TXTout, EndpointsFile, "txt")

		for _, elem := range finalEndpoints {
			for _, parameter := range elem.Parameters {
				var sb strings.Builder

				sb.WriteString(parameter.Parameter)

				if len(parameter.Attacks) > 0 {
					sb.WriteString(" -")

					for _, attack := range parameter.Attacks {
						sb.WriteString(" ")
						sb.WriteString(attack)
					}
				}

				AppendOutputToTxt(fmt.Sprintf("%s in %s", sb.String(), elem.URL), EndpointFilename)
			}
		}
	}

	// if extensions flag enabled save also secrets
	if hasValidExtensionFlag(flags.Extensions) && len(finalExtensions) > 0 {
		ExtensionsFilename := fileUtils.CreateOutputFile(flags.TXTout, ExtensionsFile, "txt")
		for _, elem := range finalExtensions {
			AppendOutputToTxt(fmt.Sprintf("%s in %s", elem.Filetype.Extension, elem.URL), ExtensionsFilename)
		}
	}

	// if errors flag enabled save also errors
	if flags.Errors && len(finalErrors) > 0 {
		ErrorsFilename := fileUtils.CreateOutputFile(flags.TXTout, ErrorsFile, "txt")
		for _, elem := range finalErrors {
			AppendOutputToTxt(fmt.Sprintf("%s - %s in %s", elem.Error.ErrorName, elem.Match, elem.URL), ErrorsFilename)
		}
	}

	// if info flag enabled save also infos
	if flags.Info && len(finalInfos) > 0 {
		InfosFilename := fileUtils.CreateOutputFile(flags.TXTout, InfosFile, "txt")
		for _, elem := range finalInfos {
			AppendOutputToTxt(fmt.Sprintf("%s - %s in %s", elem.Info.Name, elem.Match, elem.URL), InfosFilename)
		}
	}
}

// HtmlOutput it's the wrapper around all the html things.
// Actually it manages everything related to HTML output.
func HTMLOutput(flags input.Input, resultFilename string, finalResults []string, finalSecret []scanner.SecretMatched,
	finalEndpoints []scanner.EndpointMatched, finalExtensions []scanner.FileTypeMatched,
	finalErrors []scanner.ErrorMatched, finalInfos []scanner.InfoMatched) {
	exists, err := fileUtils.ElementExists(CariddiOutputFolder)
	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}

	if !exists {
		fileUtils.CreateOutputFolder()
	}

	HeaderHTML("Results found", resultFilename)

	for _, elem := range finalResults {
		AppendOutputToHTML(elem, "", resultFilename, true)
	}

	FooterHTML(resultFilename)

	// if secrets flag enabled save also secrets
	if flags.Secrets && len(finalSecret) > 0 {
		HeaderHTML("Secrets found", resultFilename)

		for _, elem := range finalSecret {
			AppendOutputToHTML(fmt.Sprintf("%s - %s in %s", elem.Secret.Name, elem.Match, elem.URL), "", resultFilename, false)
		}

		FooterHTML(resultFilename)
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints && len(finalEndpoints) > 0 {
		HeaderHTML("Endpoints found", resultFilename)

		for _, elem := range finalEndpoints {
			for _, parameter := range elem.Parameters {
				finalString := "" + parameter.Parameter
				if len(parameter.Attacks) != 0 {
					finalString += " -"
					for _, attack := range parameter.Attacks {
						finalString += " " + attack
					}
				}

				AppendOutputToHTML(fmt.Sprintf("%s in %s", finalString, elem.URL), "", resultFilename, false)
			}
		}

		FooterHTML(resultFilename)
	}

	// if extensions flag enabled save also extensions
	if hasValidExtensionFlag(flags.Extensions) && len(finalExtensions) > 0 {
		HeaderHTML("Extensions found", resultFilename)

		for _, elem := range finalExtensions {
			AppendOutputToHTML(fmt.Sprintf("%s in %s", elem.Filetype.Extension, elem.URL), "", resultFilename, false)
		}

		FooterHTML(resultFilename)
	}

	// if errors flag enabled save also errors
	if flags.Errors && len(finalErrors) > 0 {
		HeaderHTML("Errors found", resultFilename)

		for _, elem := range finalErrors {
			output := fmt.Sprintf("%s - %s in %s", elem.Error.ErrorName, elem.Match, elem.URL)
			AppendOutputToHTML(output, "", resultFilename, false)
		}

		FooterHTML(resultFilename)
	}

	// if info flag enabled save also infos
	if flags.Info && len(finalInfos) > 0 {
		HeaderHTML("Useful informations found", resultFilename)

		for _, elem := range finalInfos {
			// Escape HTML comment to be shown on the result page
			AppendOutputToHTML(elem.Info.Name+" - "+
				html.EscapeString(elem.Match)+
				" in "+elem.URL, "", resultFilename, false)
		}

		FooterHTML(resultFilename)
	}

	BannerFooterHTML(resultFilename)
}
