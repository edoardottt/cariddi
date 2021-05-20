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
func TxtOutput(flags input.Input, finalResult []string, finalSecret []scanner.SecretMatched) {
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

	ResultFilename := utils.CreateOutputFile(flags.Txt, "results", "txt")

	for _, elem := range finalResult {
		AppendOutputToTxt(elem, ResultFilename)
	}
}

//HtmlOutput it's the wrapper around all the html things.
//Actually it manages everything related to HTML output.
func HtmlOutput(flags input.Input, finalResult []string, finalSecret []scanner.SecretMatched) {
	exists, err := utils.ElementExists("output-cariddi")

	if err != nil {
		fmt.Println("Error while creating the output directory.")
		os.Exit(1)
	}

	if !exists {
		utils.CreateOutputFolder()
	}

	ResultFilename := utils.CreateOutputFile(flags.Html, "", "html")

	BannerHTML(ResultFilename)

	// if secrets flag enabled save also secrets
	if flags.Secrets {
		HeaderHTML("Secrets found", ResultFilename)
		for _, elem := range finalSecret {
			AppendOutputToHTML(elem.Secret.Name+" Found in "+elem.Url+" "+elem.Secret.Regex, "", ResultFilename, false)
		}
		FooterHTML(ResultFilename)
	}

	HeaderHTML("Results", ResultFilename)
	for _, elem := range finalResult {
		AppendOutputToHTML(elem, "", ResultFilename, true)
	}
	FooterHTML(ResultFilename)

	BannerFooterHTML(ResultFilename)
}
