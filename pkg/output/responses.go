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

package output

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gocolly/colly/v2"

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
	// The path and the query, as sent.
	builder.WriteString(resp.Request.URL.RequestURI())
	builder.WriteString(" ")
	builder.WriteString("HTTP/1.1")
	builder.WriteString("\n")
	builder.WriteString("Host: ")
	builder.WriteString(resp.Request.URL.Host)
	builder.WriteRune('\n')

	for k, v := range *resp.Request.Headers {
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(strings.Join(v, "; "))
		builder.WriteString("\n")
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
	fmt.Fprint(builder, resp.StatusCode)
	builder.WriteString("\n")

	for k, v := range *resp.Headers {
		builder.WriteString(k)
		builder.WriteString(": ")
		builder.WriteString(strings.Join(v, "; "))
		builder.WriteString("\n")
	}

	builder.WriteString("\n")
	builder.Write(resp.Body)

	return builder.Bytes(), nil
}

func getResponseFileName(folder, url string) string {
	file := getResponseHash(url) + ".txt"
	return filepath.Join(folder, file)
}

// hostFolder returns the folder of the responses of the host of u.
// The ":" of a port or of an IPv6 address is replaced, as Windows
// doesn't allow it in file names.
func hostFolder(u *url.URL) string {
	return filepath.Join(CariddiOutputFolder, strings.ReplaceAll(u.Host, ":", "_"))
}

// UpdateIndex updates the index file with the
// correct information linking to HTTP responses files.
// If it fails returns an error.
func UpdateIndex(resp *colly.Response) error {
	index, err := os.OpenFile(filepath.Join(CariddiOutputFolder, index),
		os.O_CREATE|os.O_APPEND|os.O_WRONLY,
		fileUtils.Permission0644)
	if err != nil {
		return err
	}

	defer index.Close()

	builder := &bytes.Buffer{}

	builder.WriteString(getResponseFileName(hostFolder(resp.Request.URL), resp.Request.URL.String()))
	builder.WriteRune(' ')
	builder.WriteString(resp.Request.URL.String())
	builder.WriteRune(' ')
	builder.WriteString("(")
	fmt.Fprint(builder, resp.StatusCode)
	builder.WriteString(")")
	builder.WriteRune('\n')

	if _, err := index.Write(builder.Bytes()); err != nil {
		return fmt.Errorf("could not update index: %w", err)
	}

	return nil
}

// WriteHTTPResponse creates an HTTP response output file and
// writes the HTTP response inside it (replacing the previous one).
// If it fails returns an error.
func WriteHTTPResponse(inputURL *url.URL, response []byte) error {
	file := getResponseFileName(hostFolder(inputURL), inputURL.String())

	if err := os.WriteFile(file, response, fileUtils.Permission0644); err != nil {
		return fmt.Errorf("%w: %w", ErrHTTPResp, err)
	}

	return nil
}

// StoreHTTPResponse stores an HTTP response in a file and lists it in
// the index. If it fails returns an error.
func StoreHTTPResponse(r *colly.Response) error {
	if err := os.MkdirAll(hostFolder(r.Request.URL), fileUtils.Permission0755); err != nil {
		return fmt.Errorf("%w: %w", ErrHTTPResp, err)
	}

	response, err := FormatResponse(r)
	if err != nil {
		return err
	}

	// The response file first: the index must list only stored responses.
	if err := WriteHTTPResponse(r.Request.URL, response); err != nil {
		return err
	}

	return UpdateIndex(r)
}
