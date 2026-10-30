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

package crawler

import (
	"net/url"
	"regexp"
	"strings"
	"sync"

	urlUtils "github.com/edoardottt/cariddi/v2/internal/url"
	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

// newSecretScanner returns the scanner for the predefined secret regexes or,
// if a secrets file is used, for its regexes (invalid regexes are skipped).
// Blank lines are ignored: an empty regex would match everywhere.
func newSecretScanner(secretsFile []string) *scanner.SecretScanner {
	rawRegexes := make([]string, 0, len(secretsFile))

	for _, line := range secretsFile {
		if strings.TrimSpace(line) != "" {
			rawRegexes = append(rawRegexes, line)
		}
	}

	if len(rawRegexes) == 0 {
		return scanner.GetSecretScanner()
	}

	secrets := make([]scanner.Secret, 0, len(rawRegexes))

	for _, rawRegex := range rawRegexes {
		re, err := regexp.Compile(rawRegex)
		if err != nil {
			continue // skip invalid regex
		}

		secrets = append(secrets, scanner.Secret{
			Name:        "CustomFromFile",
			Description: "",
			Regex:       *re,
			Poc:         "",
		})
	}

	return scanner.NewSecretScanner(secrets)
}

// huntSecrets hunts for secrets.
func huntSecrets(target, body string, secretScanner *scanner.SecretScanner) []scanner.SecretMatched {
	return scanner.RemoveDuplicateSecrets(secretScanner.Scan(target, body))
}

// huntEndpoints hunts for juicy endpoints.
func huntEndpoints(target string, endpointsFile *[]string) []scanner.EndpointMatched {
	endpoints := EndpointsMatch(target, endpointsFile)
	return endpoints
}

// EndpointsMatch check if an endpoint matches a juicy parameter.
func EndpointsMatch(target string, endpointsFile *[]string) []scanner.EndpointMatched {
	endpoints := []scanner.EndpointMatched{}
	matched := []scanner.Parameter{}
	parameters := urlUtils.RetrieveParameters(target)

	if len(*endpointsFile) == 0 {
		for _, parameter := range scanner.GetJuicyParameters() {
			for _, param := range parameters {
				if strings.ToLower(param) == parameter.Parameter {
					matched = append(matched, parameter)
				}
			}
		}

		endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, URL: target})
	} else {
		for _, parameter := range *endpointsFile {
			for _, param := range parameters {
				if param == parameter {
					matched = append(matched, scanner.Parameter{Parameter: parameter, Attacks: []string{}})
				}
			}
		}

		endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, URL: target})
	}

	return endpoints
}

// huntExtensions hunts for extensions in the path of u: decoded and
// without the query and the fragment (e.g. /a%2Esql#top is /a.sql).
func huntExtensions(u *url.URL, severity int) scanner.FileTypeMatched {
	urlPath := strings.ToLower(u.Path)

	for _, ext := range scanner.GetExtensions() {
		if ext.Severity <= severity && strings.HasSuffix(urlPath, "."+ext.Extension) {
			return scanner.FileTypeMatched{Filetype: ext, URL: u.String()}
		}
	}

	return scanner.FileTypeMatched{}
}

// huntErrors hunts for errors.
func huntErrors(target, body string) []scanner.ErrorMatched {
	return scanner.RemoveDuplicateErrors(scanner.GetErrorScanner().Scan(target, body))
}

// huntInfos hunts for infos.
func huntInfos(target, body string) []scanner.InfoMatched {
	var (
		infosSlice []scanner.InfoMatched
		wg         sync.WaitGroup
		mu         sync.Mutex
	)

	for _, infoItem := range scanner.GetInfoRegexes() {
		wg.Add(1)

		go func(infoItem scanner.Info) {
			defer wg.Done()

			matches := infoItem.FindAll(body)

			for _, match := range matches {
				mu.Lock()

				infosSlice = append(infosSlice, scanner.InfoMatched{Info: infoItem, URL: target, Match: match})

				mu.Unlock()
			}
		}(infoItem)
	}

	wg.Wait()

	return scanner.RemoveDuplicateInfos(infosSlice)
}
