package main

import (
	"fmt"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
)

func main() {
	fmt.Println("cariddi")
	input.ScanInput()
	output.Beautify()
}
