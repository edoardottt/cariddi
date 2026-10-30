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

	@Author:      edoardottt, https://edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package utils

import (
	"bufio"
	"fmt"
	"log"
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

// ReplaceBadCharacterOutput replaces forward-slashes
// with dashes (to avoid problems with output folder).
func ReplaceBadCharacterOutput(input string) string {
	result := strings.ReplaceAll(input, "/", "-")
	return result
}

// ReadFile reads a file line per line
// and returns a slice of strings.
// Blank lines are skipped (e.g. in an ignore
// file a blank line would match every URL).
func ReadFile(inputFile string) []string {
	file, err := os.Open(inputFile)
	if err != nil {
		log.Fatalf("failed to open %s ", inputFile)
	}

	scanner := bufio.NewScanner(file)

	var text []string

	for scanner.Scan() {
		if line := scanner.Text(); strings.TrimSpace(line) != "" {
			text = append(text, line)
		}
	}

	file.Close()

	// E.g. a line too long or a directory: the lines read so far
	// are not the whole file.
	if err := scanner.Err(); err != nil {
		log.Fatalf("failed to read %s: %v", inputFile, err)
	}

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

// ReadEntireFile returns the content of the inputted file.
func ReadEntireFile(inputFile string) []byte {
	b, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Println("Cannot read input file:", err)
		os.Exit(1)
	}

	return b
}
