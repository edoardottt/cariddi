package utils

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

//CreateOutputFolder
func CreateOutputFolder() {
	//Create a folder/directory at a full qualified path
	err := os.Mkdir("output-cariddi", 0755)
	if err != nil {
		fmt.Println("Can't create output folder.")
		os.Exit(1)
	}
}

//CreateOutputFile
func CreateOutputFile(target string, subcommand string, format string) string {
	target = ReplaceBadCharacterOutput(target)
	var filename string
	if subcommand != "" {
		filename = "output-cariddi" + "/" + target + "." + subcommand + "." + format
	} else {
		filename = "output-cariddi" + "/" + target + "." + format
	}
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		if _, err := os.Stat("output-cariddi/"); os.IsNotExist(err) {
			CreateOutputFolder()
		}
		// If the file doesn't exist, create it.
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Can't create output file.")
			os.Exit(1)
		}
		f.Close()
	} else {
		// The file already exists, check what the user want.
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("The output file already esists, do you want to overwrite? (Y/n): ")
		text, _ := reader.ReadString('\n')
		answer := strings.ToLower(text)
		answer = strings.TrimSpace(answer)

		if answer == "y" || answer == "yes" || answer == "" {
			f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("Can't create output file.")
				os.Exit(1)
			}
			err = f.Truncate(0)
			if err != nil {
				fmt.Println("Can't create output file.")
				os.Exit(1)
			}
			f.Close()
		} else {
			os.Exit(1)
		}
	}
	return filename
}

//ReplaceBadCharacterOutput --> from / to -
func ReplaceBadCharacterOutput(input string) string {
	result := strings.ReplaceAll(input, "/", "-")
	return result
}

//ReadFile reads a file line per line
//and returns a slice of strings.
func ReadFile(inputFile string) []string {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("failed to open %s ", inputFile)
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var text []string
	for scanner.Scan() {
		text = append(text, scanner.Text())
	}
	file.Close()
	return text
}

//ElementExists returns whether the given file or directory exists
func ElementExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
