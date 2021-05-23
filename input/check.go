package input

import (
	"fmt"
	"os"
	"strings"
)

//CheckDataPost
func CheckDataPost(input string) (map[string]string, error) {

	// ===== TODO =======
	return map[string]string{}, nil
}

//CheckOutputFile
func CheckOutputFile(input string) bool {
	invalid := []string{"\\", "/", "'", "\""}
	for _, elem := range invalid {
		if strings.ContainsAny(input, elem) {
			return false
		}
	}
	return true
}

//CheckFlags
func CheckFlags(flags Input) {
	if flags.Txt != "" {
		if !CheckOutputFile(flags.Txt) {
			fmt.Println("The output file must avoid weird symbols. Try to use - , _ , . instead.")
			os.Exit(1)
		}
	}

	if flags.Html != "" {
		if !CheckOutputFile(flags.Html) {
			fmt.Println("The output file must avoid weird symbols. Try to use - , _ , . instead.")
			os.Exit(1)
		}
	}
}
