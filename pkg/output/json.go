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
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/edoardottt/cariddi/pkg/scanner"
	"github.com/gocolly/colly"
)

type jsonData struct {
	URL           string         `json:"url"`
	Method        string         `json:"method"`
	StatusCode    int            `json:"status_code"`
	Words         int            `json:"words"`
	Lines         int            `json:"lines"`
	ContentType   string         `json:"content_type,omitempty"`
	ContentLength int            `json:"content_length,omitempty"`
	Matches       MatcherResults `json:"matches,omitempty"`
	// Host          string `json:"host"` # TODO: Add when migrating to Colly 2.x
}

type MatcherResults struct {
	FileType   scanner.FileType    `json:"filetype,omitempty"`
	Parameters []scanner.Parameter `json:"parameters,omitempty"`
	Errors     []MatcherResult     `json:"errors,omitempty"`
	Infos      []MatcherResult     `json:"infos,omitempty"`
	Secrets    []MatcherResult     `json:"secrets,omitempty"`
}

type MatcherResult struct {
	Name  string `json:"name"`
	Match string `json:"match"`
}

func GetJsonString(
	r *colly.Response,
	secrets []scanner.SecretMatched,
	parameters []scanner.Parameter,
	filetype scanner.FileType,
	errors []scanner.ErrorMatched,
	infos []scanner.InfoMatched,
) ([]byte, error) {

	// Parse response headers
	headers := r.Headers
	contentTypes := (*headers)["Content-Type"]
	contentType := ""
	if len(contentTypes) > 0 {
		contentType = contentTypes[0]
	}

	contentLength := 0
	contentLengths := (*headers)["Content-Length"]
	if len(contentLengths) > 0 {
		ret, err := strconv.Atoi(contentLengths[0])
		if err != nil {
			return nil, err
		}
		contentLength = ret
	}

	// Parse words from body
	words := len(strings.Fields(string(r.Body)))

	// Parse lines from body
	lines := len(strings.Split(string(r.Body), "\n"))

	// Process secrets
	secretList := []MatcherResult{}
	for _, secret := range secrets {
		secretList = append(secretList, MatcherResult{secret.Secret.Name, secret.Match})
	}

	// Process infos
	infoList := []MatcherResult{}
	for _, info := range infos {
		infoList = append(infoList, MatcherResult{info.Info.Name, info.Match})
	}

	// Process
	errorList := []MatcherResult{}
	for _, error := range errors {
		errorList = append(errorList, MatcherResult{error.Error.ErrorName, error.Match})
	}

	// Construct JSON response
	data := MatcherResults{
		FileType:   filetype,
		Parameters: parameters,
		Errors:     errorList,
		Infos:      infoList,
		Secrets:    secretList,
	}
	resp := &jsonData{
		URL:           r.Request.URL.String(),
		Method:        r.Request.Method,
		StatusCode:    r.StatusCode,
		Words:         words,
		Lines:         lines,
		ContentType:   contentType,
		ContentLength: contentLength,
		Matches:       data,
		// Host:          "", // TODO: this is available in Colly 2.x but not in 1.2
	}

	// Convert struct to JSON string
	jsonOutput, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	// Output JSON string
	fmt.Println(string(jsonOutput))

	return jsonOutput, nil
}