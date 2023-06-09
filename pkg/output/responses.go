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

	@Author:      edoardottt, https://www.edoardoottavianelli.it

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package output

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly"

	fileUtils "github.com/edoardottt/cariddi/internal/file"
)

const (
	index = "index.responses.txt"
)

var (
	ErrHTTPResp = errors.New("cannot store HTTP response")
)

func getResponseHash(url string) string {
	hash := sha1.Sum([]byte(url))
	return hex.EncodeToString(hash[:])
}

// FormatResponse formats an HTTP response ready to be written in a file.
func FormatResponse(resp *colly.Response) ([]byte, error) {
	builder := &bytes.Buffer{}

	builder.WriteString(resp.Request.URL.String())
	builder.WriteString("\n\n\n")

	builder.WriteString(resp.Request.Method)
	builder.WriteString(" ")

	path := resp.Request.URL.Path
	if resp.Request.URL.Fragment != "" {
		path = path + "#" + resp.Request.URL.Fragment
	}

	builder.WriteString(path)
	builder.WriteString(" ")
	builder.WriteString("HTTP/1.1")
	builder.WriteString("\n")
	builder.WriteString("Host: " + resp.Request.URL.Host)
	builder.WriteRune('\n')

	for k, v := range *resp.Request.Headers {
		builder.WriteString(k + ": " + strings.Join(v, "; ") + "\n")
	}

	if resp.Request.Body != nil {
		bodyResp, _ := io.ReadAll(resp.Request.Body)
		if string(bodyResp) != "" {
			builder.WriteString("\n")
			builder.WriteString(string(bodyResp))
		}
	}

	builder.WriteString("\n\n")
	builder.WriteString("HTTP/1.1")
	builder.WriteString(" ")
	builder.WriteString(fmt.Sprint(resp.StatusCode))
	builder.WriteString("\n")

	for k, v := range *resp.Headers {
		builder.WriteString(k + ": " + strings.Join(v, "; ") + "\n")
	}

	builder.WriteString("\n")

	body, _ := io.ReadAll(bytes.NewReader(resp.Body))

	builder.WriteString(string(body))

	return builder.Bytes(), nil
}

func getResponseFileName(folder, url string) string {
	file := getResponseHash(url) + ".txt"
	return filepath.Join(folder, file)
}

// UpdateIndex updates the index file with the
// correct information linking to HTTP responses files.
// If it fails returns an error.
func UpdateIndex(resp *colly.Response) error {
	index, err := os.OpenFile(filepath.Join(CariddiOutputFolder, index),
		os.O_APPEND|os.O_WRONLY,
		fileUtils.Permission0644)
	if err != nil {
		return err
	}

	defer index.Close()

	builder := &bytes.Buffer{}

	builder.WriteString(getResponseFileName(filepath.Join(CariddiOutputFolder, resp.Request.URL.Host),
		resp.Request.URL.String()))
	builder.WriteRune(' ')
	builder.WriteString(resp.Request.URL.String())
	builder.WriteRune(' ')
	builder.WriteString("(" + fmt.Sprint(resp.StatusCode) + ")")
	builder.WriteRune('\n')

	if _, writeErr := index.Write(builder.Bytes()); writeErr != nil {
		return fmt.Errorf("%w %s", err, "could not update index")
	}

	return nil
}

// WriteHTTPResponse creates an HTTP response output file and
// writes the HTTP response inside it.
// If it fails returns an error.
func WriteHTTPResponse(inputURL *url.URL, response []byte) error {
	file := getResponseFileName(filepath.Join(CariddiOutputFolder, inputURL.Host), inputURL.String())

	outFile, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, fileUtils.Permission0644)
	if err != nil {
		return err
	}

	if _, writeErr := outFile.Write(response); writeErr != nil {
		return ErrHTTPResp
	}

	return nil
}

// StoreHTTPResponse stores an HTTP response in a file.
// If it fails returns an error.
func StoreHTTPResponse(r *colly.Response) error {
	fileUtils.CreateHostOutputFolder(r.Request.URL.Host)

	err := UpdateIndex(r)
	if err != nil {
		log.Println(err)
	}

	response, err := FormatResponse(r)
	if err != nil {
		log.Println(err)
	}

	err = WriteHTTPResponse(r.Request.URL, response)
	if err != nil {
		log.Println(err)
	}

	return nil
}
