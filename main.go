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
	fmt.Println("cariddi")
	input := input.ScanInput()
	for _, inp := range input {
		crawler.Crawler(inp)
	}
	output.Beautify()
	scanner.Scan()
}
