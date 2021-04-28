package output

import "fmt"

//PrintOutput
func PrintOutput(out []string) {
	for _, elem := range out {
		fmt.Println(elem)
	}
}
