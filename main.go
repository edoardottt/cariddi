package main

import (
	"fmt"

	"github.com/edoardottt/cariddi/crawler"
	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
	"github.com/edoardottt/cariddi/scanner"
)

//main
func main() {
	targets := input.ScanTargets()
	flags := input.ScanFlag()
	fmt.Println(flags)
	var finalResult []string
	for _, inp := range targets {
		var result []string
		result = crawler.Crawler(inp)
		finalResult = append(finalResult, result...)
	}
	output.PrintSimpleOutput(finalResult)
	scanner.Scan()
	fmt.Println(flags)
}
