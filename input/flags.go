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
	"flag"
)

//Input
type Input struct {
	//Verbose     bool
	Version     bool
	Delay       int
	Concurrency int
	Help        bool
	Examples    bool
	Plain       bool
	Html        string
	Txt         string
	//DataPost    string
	Secrets       bool
	SecretsFile   string
	Endpoints     bool
	EndpointsFile string
	Extensions    int
}

//ScanFlag defines all the switches taken
//as input and return them.
func ScanFlag() Input {

	//verbosePtr := flag.Bool("v", false, "Verbose mode.")
	versionPtr := flag.Bool("version", false, "Print the version.")
	delayPtr := flag.Int("d", 0, "Delay between a page crawled and another.")
	concurrencyPtr := flag.Int("c", 20, "Concurrency level.")
	helpPtr := flag.Bool("h", false, "Print the help.")
	examplesPtr := flag.Bool("examples", false, "Print the examples.")
	plainPtr := flag.Bool("plain", false, "Print only the results.")
	outputHtmlPtr := flag.String("oh", "", "Write the output into an HTML file.")
	outputTxtPtr := flag.String("ot", "", "Write the output into a TXT file.")

	//dataPostPtr := flag.String("post", "", "Set the data to perform the POST requests.")

	secretsPtr := flag.Bool("s", false, "Hunt for secrets.")
	secretsFilePtr := flag.String("sf", "", "Use an external file (txt, one per line) to use custom regexes for secrets hunting.")

	endpointsPtr := flag.Bool("e", false, "Hunt for juicy endpoints.")
	endpointsFilePtr := flag.String("ef", "", "Use an external file (txt, one per line) to use custom parameters for endpoints hunting.")

	extensionsPtr := flag.Int("ext", 0, "Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).")

	flag.Parse()

	result := Input{
		//*verbosePtr,
		*versionPtr,
		*delayPtr,
		*concurrencyPtr,
		*helpPtr,
		*examplesPtr,
		*plainPtr,
		*outputHtmlPtr,
		*outputTxtPtr,
		//*dataPostPtr,
		*secretsPtr,
		*secretsFilePtr,
		*endpointsPtr,
		*endpointsFilePtr,
		*extensionsPtr,
	}

	return result
}
