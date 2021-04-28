package output

import (
	"fmt"
)

//PrintOutput
func PrintSimpleOutput(out []string) {
	for _, elem := range out {
		fmt.Println(elem)
	}
}
