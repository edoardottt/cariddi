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

package output

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"

	fileUtils "github.com/edoardottt/cariddi/v2/internal/file"
	"github.com/edoardottt/cariddi/v2/pkg/input"
	"github.com/edoardottt/cariddi/v2/pkg/scanner"
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

		lines := make([]string, 0, len(finalSecret))
		for _, elem := range finalSecret {
			lines = append(lines, fmt.Sprintf("%s - %s in %s", elem.Secret.Name, elem.Match, elem.URL))
		}

		WriteAllTxt(lines, SecretFilename)
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints && len(finalEndpoints) > 0 {
		EndpointFilename := fileUtils.CreateOutputFile(flags.TXTout, EndpointsFile, "txt")

		var lines []string

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

				lines = append(lines, fmt.Sprintf("%s in %s", sb.String(), elem.URL))
			}
		}

		WriteAllTxt(lines, EndpointFilename)
	}

	// if extensions flag enabled save also secrets
	if hasValidExtensionFlag(flags.Extensions) && len(finalExtensions) > 0 {
		ExtensionsFilename := fileUtils.CreateOutputFile(flags.TXTout, ExtensionsFile, "txt")

		lines := make([]string, 0, len(finalExtensions))
		for _, elem := range finalExtensions {
			lines = append(lines, fmt.Sprintf("%s in %s", elem.Filetype.Extension, elem.URL))
		}

		WriteAllTxt(lines, ExtensionsFilename)
	}

	// if errors flag enabled save also errors
	if flags.Errors && len(finalErrors) > 0 {
		ErrorsFilename := fileUtils.CreateOutputFile(flags.TXTout, ErrorsFile, "txt")

		lines := make([]string, 0, len(finalErrors))
		for _, elem := range finalErrors {
			lines = append(lines, fmt.Sprintf("%s - %s in %s", elem.Error.ErrorName, elem.Match, elem.URL))
		}

		WriteAllTxt(lines, ErrorsFilename)
	}

	// if info flag enabled save also infos
	if flags.Info && len(finalInfos) > 0 {
		InfosFilename := fileUtils.CreateOutputFile(flags.TXTout, InfosFile, "txt")

		lines := make([]string, 0, len(finalInfos))
		for _, elem := range finalInfos {
			lines = append(lines, fmt.Sprintf("%s - %s in %s", elem.Info.Name, elem.Match, elem.URL))
		}

		WriteAllTxt(lines, InfosFilename)
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

	// Every card with at least one finding links to its section.
	report := htmlReport{
		Generated: time.Now().Format(time.DateTime),
		Stats: []htmlStat{
			{"results", "Results", len(finalResults)},
			{"secrets", "Secrets", len(finalSecret)},
			{"endpoints", "Endpoints", len(finalEndpoints)},
			{"extensions", "Extensions", len(finalExtensions)},
			{"errors", "Errors", len(finalErrors)},
			{"info", "Info", len(finalInfos)},
		},
		Results: finalResults,
	}

	// if secrets flag enabled save also secrets
	if flags.Secrets {
		report.Secrets = finalSecret
	}

	// if endpoints flag enabled save also endpoints
	if flags.Endpoints {
		report.Endpoints = finalEndpoints
	}

	// if extensions flag enabled save also extensions (juiciest first)
	if hasValidExtensionFlag(flags.Extensions) {
		// Sort a copy: the caller still prints the extensions in crawl order.
		report.Extensions = slices.Clone(finalExtensions)
		slices.SortStableFunc(report.Extensions, func(a, b scanner.FileTypeMatched) int {
			return cmp.Or(cmp.Compare(a.Filetype.Severity, b.Filetype.Severity),
				strings.Compare(a.Filetype.Extension, b.Filetype.Extension))
		})
	}

	// if errors flag enabled save also errors
	if flags.Errors {
		report.Errors = finalErrors
	}

	// if info flag enabled save also infos
	if flags.Info {
		report.Infos = finalInfos
	}

	if err := writeHTML(resultFilename, report); err != nil {
		log.Fatal(err)
	}
}
