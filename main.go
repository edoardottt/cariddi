package main

import (
	"fmt"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
)

//main
func main() {
	fmt.Println("cariddi")
	input.ScanInput()
	output.Beautify()
}
