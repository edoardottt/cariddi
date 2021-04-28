package output

import (
	"fmt"

	"github.com/fatih/color"
)

//EncapsulateGreen
func EncapsulateGreen(inp string) {
	// Create a custom print function for convenience
	green := color.New(color.FgGreen).PrintfFunc()
	green("[ + ] ")
	fmt.Println(inp)
}

//EncapsulateRed
func EncapsulateRed(inp string) {
	// Create a custom print function for convenience
	red := color.New(color.FgRed).PrintfFunc()
	red("[ - ] ")
	fmt.Println(inp)
}

//EncapsulateYellow
func EncapsulateYellow(inp string) {
	// Create a custom print function for convenience
	yellow := color.New(color.FgYellow).PrintfFunc()
	yellow("[ ? ] ")
	fmt.Println(inp)
}

//EncapsulateGreen
func EncapsulateCustomGreen(alert string, inp string) {
	// Create a custom print function for convenience
	green := color.New(color.FgGreen).PrintfFunc()
	green("[ %s ] ", alert)
	fmt.Println(inp)
}

//EncapsulateRed
func EncapsulateCustomRed(alert string, inp string) {
	// Create a custom print function for convenience
	red := color.New(color.FgRed).PrintfFunc()
	red("[ %s ] ", alert)
	fmt.Println(inp)
}

//EncapsulateYellow
func EncapsulateCustomYellow(alert string, inp string) {
	// Create a custom print function for convenience
	yellow := color.New(color.FgYellow).PrintfFunc()
	yellow("[ %s ] ", alert)
	fmt.Println(inp)
}
