/*
==========
Cariddi
==========

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see http://www.gnu.org/licenses/.

	@Repository:  https://github.com/edoardottt/cariddi

	@Author:      edoardottt, https://www.edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package utils

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	Permission0755 = 0755
	Permission0644 = 0644
)

// CreateOutputFolder creates the output folder
// If it fails exits with an error message.
func CreateOutputFolder() {
	// Create a folder/directory at a full qualified path
	err := os.Mkdir("output-cariddi", Permission0755)
	if err != nil {
		fmt.Println("Can't create output folder.")
		os.Exit(1)
	}
}

// CreateHostOutputFolder creates the host output folder
// for the HTTP responses.
// If it fails exits with an error message.
func CreateHostOutputFolder(host string) {
	// Create a folder/directory at a full qualified path
	err := os.MkdirAll(filepath.Join("output-cariddi", host), Permission0755)
	if err != nil {
		fmt.Println("Can't create host output folder.")
		os.Exit(1)
	}
}

// CreateOutputFile takes as input a target (of the attack), a subcommand
// and a format (json-html-txt).
// It creates the output folder if needed, then checks if the output file
// already exists, if yes asks the user if cariddi has to overwrite it;
// if no cariddi creates it.
// Whenever an instruction fails, it exits with an error message.
func CreateOutputFile(target string, subcommand string, format string) string {
	target = ReplaceBadCharacterOutput(target)

	var filename string
	if subcommand != "" {
		filename = filepath.Join("output-cariddi", target+"."+subcommand+"."+format)
	} else {
		filename = filepath.Join("output-cariddi", target+"."+format)
	}

	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if _, err := os.Stat("output-cariddi/"); os.IsNotExist(err) {
			CreateOutputFolder()
		}
		// If the file doesn't exist, create it.
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, Permission0644)
		if err != nil {
			fmt.Println("Can't create output file.")
			os.Exit(1)
		}

		f.Close()
	} else {
		// The file already exists, overwrite.
		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, Permission0644)
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
	}

	return filename
}

// CreateIndexOutputFile takes as input the name of the index file.
// It creates the output folder if needed, then checks if the index output file
// already exists, if no cariddi creates it.
// Whenever an instruction fails, it exits with an error message.
func CreateIndexOutputFile(filename string) {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		if _, err := os.Stat("output-cariddi/"); os.IsNotExist(err) {
			CreateOutputFolder()
		}
		// If the file doesn't exist, create it.
		filename = filepath.Join("output-cariddi", filename)

		f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, Permission0644)
		if err != nil {
			fmt.Println("Can't create output file.")
			os.Exit(1)
		}

		f.Close()
	}
}

// ReplaceBadCharacterOutput replaces forward-slashes
// with dashes (to avoid problems with output folder).
func ReplaceBadCharacterOutput(input string) string {
	result := strings.ReplaceAll(input, "/", "-")
	return result
}

// ReadFile reads a file line per line
// and returns a slice of strings.
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

// ElementExists returns whether the given file or directory exists.
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

// ReadHTTPRequestFromFile reads from a file an HTTP
// request and returns a *http.Request object.
func ReadHTTPRequestFromFile(inputFile string) (*http.Request, error) {
	f, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Cannot open input file.")
		os.Exit(1)
	}
	defer f.Close()

	buf := bufio.NewReader(f)

	req, err := http.ReadRequest(buf)
	if err != nil {
		fmt.Println("Cannot read request from input file.")
		return req, err
	}

	return req, nil
}

// ReadEntireFile returns the content of the inputted file.
func ReadEntireFile(inputFile string) []byte {
	file, err := os.Open(inputFile)
	if err != nil {
		fmt.Println("Cannot open input file.")
		os.Exit(1)
	}

	defer func() {
		if err = file.Close(); err != nil {
			fmt.Println("Cannot close input file.")
			os.Exit(1)
		}
	}()

	b, err := io.ReadAll(file)

	return b
}
