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

	var finalResult []string
	for _, inp := range targets {
		var result []string
		result = crawler.Crawler(inp, flags.Delay, flags.Concurrency, flags.Secrets, flags.SecretsFile, data)
		finalResult = append(finalResult, result...)
	}
	output.PrintSimpleOutput(finalResult)
	scanner.Scan()
}
