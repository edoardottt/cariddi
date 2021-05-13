package input

import "flag"

//Input
type Input struct {
	Verbose     bool
	Version     bool
	Delay       int
	Concurrency int
	Help        bool
	Examples    bool
	Html        string
	Txt         string
	DataPost    string
	Secrets     bool
	SecretsFile string
}

//ScanFlag defines all the switches taken
//as input and return them.
func ScanFlag() Input {

	verbosePtr := flag.Bool("v", false, "Verbose mode.")
	versionPtr := flag.Bool("version", false, "Print the version.")
	delayPtr := flag.Int("d", 0, "Delay between a page crawled and another.")
	concurrencyPtr := flag.Int("c", 20, "Concurrency level (20 is default).")
	helpPtr := flag.Bool("h", false, "Print the version.")
	examplesPtr := flag.Bool("examples", false, "Print the version.")
	outputHtmlPtr := flag.String("oh", "", "Write the output into an HTML file.")
	outputTxtPtr := flag.String("ot", "", "Write the output into a TXT file.")
	dataPostPtr := flag.String("post", "", "Set the data to perform the POST requests.")
	secretsPtr := flag.Bool("s", false, "Hunt for secrets.")
	secretsFilePtr := flag.String("sf", "", "Use an external file (txt, one per line) to use custom regexes for secrets hunting.")

	flag.Parse()

	result := Input{
		*verbosePtr,
		*versionPtr,
		*delayPtr,
		*concurrencyPtr,
		*helpPtr,
		*examplesPtr,
		*outputHtmlPtr,
		*outputTxtPtr,
		*dataPostPtr,
		*secretsPtr,
		*secretsFilePtr,
	}

	return result
}
