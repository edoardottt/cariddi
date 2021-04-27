package input

import (
	"bufio"
	"os"
	"strings"
)

//ScanInput return the array of elements
//taken as input on stdin.
func ScanInput() []string {

	var result []string

	// accept domains on stdin
	sc := bufio.NewScanner(os.Stdin)
	for sc.Scan() {
		domain := strings.ToLower(sc.Text())
		result = append(result, domain)
	}
	return result
}
