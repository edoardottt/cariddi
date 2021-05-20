package main

import (
	"fmt"
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

	fmt.Println("FLAGS:")
	fmt.Println(flags)
	fmt.Println("--------------")

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

	// ----------- TODO: check flags.dataPost --------------
	data, _ := input.CheckDataPost(flags.DataPost)

	// ----------- TODO: check ALL input -------------------

	output.Beautify()

	var finalResult []string
	var finalSecret []scanner.SecretMatched
	for _, inp := range targets {
		result, secrets := crawler.Crawler(inp, flags.Delay, flags.Concurrency, flags.Secrets, flags.SecretsFile, flags.Plain, data)
		finalResult = append(finalResult, result...)
		finalSecret = append(finalSecret, secrets...)
	}

	// IF TXT OUTPUT
	if flags.Txt != "" {

		// if secrets flag enabled save also secrets
		if flags.Secrets {

		}

	}

	// IF HTML OUTPUT
	if flags.Html != "" {

		// if secrets flag enabled save also secrets
		if flags.Secrets {

		}

	}

	// if needed print secrets
	if !flags.Plain {
		for _, elem := range finalSecret {
			output.EncapsulateCustomGreen(elem.Secret.Name, "Found in "+elem.Url+" "+elem.Secret.Regex+" matched!")
		}
	}
	// if needed print urls
	if !flags.Plain {
		output.PrintSimpleOutput(finalResult)
	}
}
