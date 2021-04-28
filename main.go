package main

import (
	"github.com/edoardottt/cariddi/crawler"
	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
	"github.com/edoardottt/cariddi/scanner"
)

//main
func main() {
	input := input.ScanInput()
	var finalResult []string
	for _, inp := range input {
		var result []string
		result = crawler.Crawler(inp)
		finalResult = append(finalResult, result...)
	}
	output.PrintOutput(finalResult)
	scanner.Scan()
}
