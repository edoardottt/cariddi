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
	"strconv"
	"strings"

	"github.com/edoardottt/cariddi/pkg/scanner"
	"github.com/gocolly/colly"
)

type JSONData struct {
	URL           string          `json:"url"`
	Method        string          `json:"method"`
	StatusCode    int             `json:"status_code"`
	Words         int             `json:"words"`
	Lines         int             `json:"lines"`
	ContentType   string          `json:"content_type,omitempty"`
	ContentLength int             `json:"content_length,omitempty"`
	Matches       *MatcherResults `json:"matches,omitempty"`
	// Host          string `json:"host"` # TODO: Available in Colly 2.x
}

type MatcherResults struct {
	FileType   *scanner.FileType   `json:"filetype,omitempty"`
	Parameters []scanner.Parameter `json:"parameters,omitempty"`
	Errors     map[string][]string `json:"errors,omitempty"`
	Infos      map[string][]string `json:"infos,omitempty"`
	Secrets    map[string][]string `json:"secrets,omitempty"`
}

// GetJSONString returns the JSON byte object.
func GetJSONString(
	r *colly.Response,
	secrets []scanner.SecretMatched,
	parameters []scanner.Parameter,
	filetype *scanner.FileType,
	errors []scanner.ErrorMatched,
	infos []scanner.InfoMatched,
) ([]byte, error) {
	// Parse response headers
	headers := r.Headers
	contentTypes := (*headers)["Content-Type"]
	contentLengths := (*headers)["Content-Length"]
	contentType := ""
	contentLength := 0

	// Set content type
	if len(contentTypes) > 0 {
		contentType = strings.Split(contentTypes[0], "; ")[0]
	}

	// Set content length
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

	secretM := processSecrets(secrets)
	infoM := processInfos(infos)
	errorM := processErrors(errors)

	// Construct matcher results
	matcherResults := &MatcherResults{
		FileType:   filetype,
		Parameters: parameters,
		Errors:     errorM,
		Infos:      infoM,
		Secrets:    secretM,
	}

	// Construct JSON response
	resp := &JSONData{
		URL:           r.Request.URL.String(),
		Method:        r.Request.Method,
		StatusCode:    r.StatusCode,
		Words:         words,
		Lines:         lines,
		ContentType:   contentType,
		ContentLength: contentLength,
		Matches:       matcherResults,
		// Host:          "", // TODO: this is available in Colly 2.x
	}

	// Set empty data if no matches to bridge the omitempty gap for empty structs
	var (
		isFileTypeNill    = false
		isParametersEmpty = len(parameters) == 0
		isErrorsEmpty     = len(errorM) == 0
		isInfoEmpty       = len(infoM) == 0
		isSecretsEmpty    = len(secretM) == 0
	)

	if (*filetype == scanner.FileType{}) {
		matcherResults.FileType = nil
		isFileTypeNill = true
	}

	if isFileTypeNill && isParametersEmpty && isErrorsEmpty && isInfoEmpty && isSecretsEmpty {
		resp.Matches = nil
	}

	// Convert struct to JSON string
	jsonOutput, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}

	return jsonOutput, nil
}

func processSecrets(secrets []scanner.SecretMatched) map[string][]string {
	secretM := map[string][]string{}

	for _, secret := range secrets {
		if _, ok := secretM[secret.Secret.Name]; ok {
			tempV := secretM[secret.Secret.Name]
			tempV = append(tempV, secret.Match)
			secretM[secret.Secret.Name] = tempV
		} else {
			secretM[secret.Secret.Name] = []string{secret.Match}
		}
	}

	return secretM
}

func processInfos(infos []scanner.InfoMatched) map[string][]string {
	infoM := map[string][]string{}

	for _, info := range infos {
		if _, ok := infoM[info.Info.Name]; ok {
			tempV := infoM[info.Info.Name]
			tempV = append(tempV, info.Match)
			infoM[info.Info.Name] = tempV
		} else {
			infoM[info.Info.Name] = []string{info.Match}
		}
	}

	return infoM
}

func processErrors(errors []scanner.ErrorMatched) map[string][]string {
	errorM := map[string][]string{}

	for _, er := range errors {
		if _, ok := errorM[er.Error.ErrorName]; ok {
			tempV := errorM[er.Error.ErrorName]
			tempV = append(tempV, er.Match)
			errorM[er.Error.ErrorName] = tempV
		} else {
			errorM[er.Error.ErrorName] = []string{er.Match}
		}
	}

	return errorM
}
