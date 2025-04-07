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

package crawler

import (
	"regexp"
	"strings"
	"sync"

	urlUtils "github.com/edoardottt/cariddi/internal/url"
	"github.com/edoardottt/cariddi/pkg/scanner"
)

// huntSecrets hunts for secrets.
func huntSecrets(target, body string, secretsFile *[]string) []scanner.SecretMatched {
	var (
		secrets []scanner.SecretMatched
		mu      sync.Mutex
		wg      sync.WaitGroup
	)

	if len(*secretsFile) == 0 { // Predefined secret regexes
		for _, secret := range scanner.GetSecretRegexes() {
			wg.Add(1)

			go func(secret scanner.Secret) {
				defer wg.Done()

				matches := secret.Regex.FindAllStringSubmatch(body, -1)

				for _, match := range matches {
					isFalsePositive := false

					for _, fp := range secret.FalsePositives {
						if strings.Contains(strings.ToLower(match[0]), fp) {
							isFalsePositive = true
							break
						}
					}

					if isFalsePositive {
						continue
					}

					secretMatch := scanner.SecretMatched{
						Secret: secret,
						URL:    target,
						Match:  match[0],
					}

					mu.Lock()
					secrets = append(secrets, secretMatch)
					mu.Unlock()
				}
			}(secret)
		}
	} else { // Custom regex from file
		for _, rawRegex := range *secretsFile {
			re, err := regexp.Compile(rawRegex)
			if err != nil {
				continue // skip invalid regex
			}

			wg.Add(1)

			go func(re *regexp.Regexp) {
				defer wg.Done()

				matches := re.FindAllStringSubmatch(body, -1)
				if matches == nil {
					return
				}

				for _, match := range matches {
					secretMatch := scanner.SecretMatched{
						Secret: scanner.Secret{
							Name:        "CustomFromFile",
							Description: "",
							Regex:       *re,
							Poc:         "",
						},
						URL:   target,
						Match: match[0],
					}

					mu.Lock()
					secrets = append(secrets, secretMatch)
					mu.Unlock()
				}
			}(re)
		}
	}

	wg.Wait()

	return scanner.RemoveDuplicateSecrets(secrets)
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

// huntExtensions hunts for extensions.
func huntExtensions(target string, severity int) scanner.FileTypeMatched {
	extension := scanner.FileTypeMatched{}
	copyTarget := target

	for _, ext := range scanner.GetExtensions() {
		if ext.Severity <= severity {
			firstIndex := strings.Index(target, "?")
			if firstIndex > -1 {
				target = target[:firstIndex]
			}

			if strings.ToLower(target[len(target)-len("."+ext.Extension):]) == "."+ext.Extension {
				extension = scanner.FileTypeMatched{Filetype: ext, URL: copyTarget}
			}
		}
	}

	return extension
}

// huntErrors hunts for errors.
func huntErrors(target, body string) []scanner.ErrorMatched {
	var (
		results []scanner.ErrorMatched
		wg      sync.WaitGroup
		mutex   sync.Mutex
	)

	errorRegexes := scanner.GetErrorRegexes()

	for _, err := range errorRegexes {
		wg.Add(1)

		go func(err scanner.Error) {
			defer wg.Done()

			matches := err.Regex.FindAllStringSubmatch(body, -1)
			if matches == nil {
				return
			}

			var localResults []scanner.ErrorMatched
			for _, match := range matches {
				localResults = append(localResults, scanner.ErrorMatched{
					Error: err,
					URL:   target,
					Match: match[0],
				})
			}

			mutex.Lock()
			results = append(results, localResults...)
			mutex.Unlock()
		}(err)
	}

	wg.Wait()

	return scanner.RemoveDuplicateErrors(results)
}

// huntInfos hunts for infos.
func huntInfos(target, body string) []scanner.InfoMatched {
	var (
		infosSlice []scanner.InfoMatched
		wg         sync.WaitGroup
		mu         sync.Mutex // Mutex to protect shared resource
	)

	// Iterate over each pattern in a goroutine
	for _, infoItem := range scanner.GetInfoRegexes() {
		wg.Add(1)

		go func(infoItem scanner.Info) {
			defer wg.Done()

			matches := infoItem.Regex.FindAllStringSubmatch(body, -1)

			// Process the matches
			for _, match := range matches {
				// Lock the shared resource before modifying it
				mu.Lock()
				infosSlice = append(infosSlice, scanner.InfoMatched{Info: infoItem, URL: target, Match: match[0]})
				mu.Unlock()
			}
		}(infoItem)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	// Remove duplicate infos
	return scanner.RemoveDuplicateInfos(infosSlice)
}
