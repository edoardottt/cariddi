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

	urlUtils "github.com/edoardottt/cariddi/internal/url"
	"github.com/edoardottt/cariddi/pkg/scanner"
)

// huntSecrets hunts for secrets.
func huntSecrets(target, body string, secretsFile *[]string) []scanner.SecretMatched {
	secrets := SecretsMatch(target, body, secretsFile)
	return secrets
}

// SecretsMatch checks if a body matches some secrets.
func SecretsMatch(url, body string, secretsFile *[]string) []scanner.SecretMatched {
	var secrets []scanner.SecretMatched

	if len(*secretsFile) == 0 {
		for _, secret := range scanner.GetSecretRegexes() {
			matches := secret.Regex.FindAllStringSubmatch(body, -1)

			// Avoiding false positives
			var isFalsePositive = false

			for _, match := range matches {
				for _, falsePositive := range secret.FalsePositives {
					if strings.Contains(strings.ToLower(match[0]), falsePositive) {
						isFalsePositive = true
						break
					}
				}

				if !isFalsePositive {
					secretFound := scanner.SecretMatched{Secret: secret, URL: url, Match: match[0]}
					secrets = append(secrets, secretFound)
				}
			}
		}
	} else {
		for _, secret := range *secretsFile {
			if matched, err := regexp.Match(secret, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(secret)
				matches := re.FindAllStringSubmatch(body, -1)

				for _, match := range matches {
					secretScanned := scanner.Secret{Name: "CustomFromFile",
						Description: "",
						Regex:       *regexp.MustCompile(secret),
						Poc:         ""}
					secretFound := scanner.SecretMatched{Secret: secretScanned, URL: url, Match: match[0]}
					secrets = append(secrets, secretFound)
				}
			}
		}
	}

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
	errorsSlice := ErrorsMatch(target, body)
	return errorsSlice
}

// ErrorsMatch checks the patterns for errors.
func ErrorsMatch(url, body string) []scanner.ErrorMatched {
	errors := []scanner.ErrorMatched{}

	for _, errorItem := range scanner.GetErrorRegexes() {
		matches := errorItem.Regex.FindAllStringSubmatch(body, -1)

		for _, match := range matches {
			errorFound := scanner.ErrorMatched{Error: errorItem, URL: url, Match: match[0]}
			errors = append(errors, errorFound)
		}
	}

	return scanner.RemoveDuplicateErrors(errors)
}

// huntInfos hunts for infos.
func huntInfos(target, body string) []scanner.InfoMatched {
	infosSlice := InfoMatch(target, body)
	return infosSlice
}

// InfoMatch checks the patterns for infos.
func InfoMatch(url, body string) []scanner.InfoMatched {
	infos := []scanner.InfoMatched{}

	for _, infoItem := range scanner.GetInfoRegexes() {
		matches := infoItem.Regex.FindAllStringSubmatch(body, -1)

		for _, match := range matches {
			infoFound := scanner.InfoMatched{Info: infoItem, URL: url, Match: match[0]}
			infos = append(infos, infoFound)
		}
	}

	return scanner.RemoveDuplicateInfos(infos)
}
