package input

import (
	"bufio"
	"flag"
	"os"
	"strings"
)

//Input
type Input struct {
	Verbose     bool
	Html        string
	Txt         string
	Secrets     bool
	SecretsFile string
}

//ScanFlag
func ScanFlag() Input {

	verbosePtr := flag.Bool("v", false, "Verbose mode.")
	outputHtmlPtr := flag.String("oh", "", "Write the output into an HTML file.")
	outputTxtPtr := flag.String("ot", "", "Write the output into a TXT file.")
	secretsPtr := flag.Bool("s", false, "Hunt for secrets.")
	secretsFilePtr := flag.String("sf", "", "Use an external file (txt, one per line) to use custom regexes for secrets hunting.")

	result := Input{
		*verbosePtr,
		*outputHtmlPtr,
		*outputTxtPtr,
		*secretsPtr,
		*secretsFilePtr,
	}

	return result
}

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
