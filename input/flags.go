package input

import "flag"

//Input
type Input struct {
	Verbose     bool
	Version     bool
	Html        string
	Txt         string
	Secrets     bool
	SecretsFile string
}

//ScanFlag defines all the switches taken
//as input and return them.
func ScanFlag() Input {

	verbosePtr := flag.Bool("v", false, "Verbose mode.")
	versionPtr := flag.Bool("version", false, "Print the version.")
	outputHtmlPtr := flag.String("oh", "", "Write the output into an HTML file.")
	outputTxtPtr := flag.String("ot", "", "Write the output into a TXT file.")
	secretsPtr := flag.Bool("s", false, "Hunt for secrets.")
	secretsFilePtr := flag.String("sf", "", "Use an external file (txt, one per line) to use custom regexes for secrets hunting.")

	result := Input{
		*verbosePtr,
		*versionPtr,
		*outputHtmlPtr,
		*outputTxtPtr,
		*secretsPtr,
		*secretsFilePtr,
	}

	return result
}
