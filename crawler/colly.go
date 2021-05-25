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
*/

package crawler

import (
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/gocolly/colly"
)

//Crawler
func Crawler(target string, delayTime int, concurrency int, secrets bool, secretsFile []string, plain bool, endpoints bool, endpointsFile []string, fileType int) ([]string, []scanner.SecretMatched, []scanner.EndpointMatched, []scanner.FileTypeMatched) {

	//clean target input
	target = input.RemoveHeaders(target)

	var Finalresult []string
	var Finalsecrets []scanner.SecretMatched
	var Finalendpoints []scanner.EndpointMatched
	var FinalExtensions []scanner.FileTypeMatched

	// Instantiate  collector
	c := colly.NewCollector(
		colly.AllowedDomains(target),
		colly.Async(true),
		colly.URLFilters(
			regexp.MustCompile(target+"*"),
		),
	)

	c.Limit(
		&colly.LimitRule{
			DomainGlob:  target,
			Parallelism: concurrency,
			Delay:       time.Duration(delayTime) * time.Second,
		},
	)
	c.AllowURLRevisit = false

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// On every a script element which has src attribute call callback
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	c.OnResponse(func(r *colly.Response) {
		// HERE SCAN FOR SECRETS
		if secrets {
			// DON'T SCAN THE URLS SCANNED BEFORE
			secretsSlice := huntSecrets(secretsFile, r.Request.URL.String(), string(r.Body))
			for _, elem := range secretsSlice {

				secretFound := scanner.SecretMatched{Secret: elem, Url: r.Request.URL.String()}
				Finalsecrets = append(Finalsecrets, secretFound)
			}
		}
		// HERE SCAN FOR ENDPOINTS
		if endpoints {
			// DON'T SCAN THE URLS SCANNED BEFORE
			endpointsSlice := huntEndpoints(endpointsFile, r.Request.URL.String())
			for _, elem := range endpointsSlice {
				if len(elem.Parameters) != 0 {
					Finalendpoints = append(Finalendpoints, elem)
				}
			}
		}
		// HERE SCAN FOR EXTENSIONS
		if 1 <= fileType && fileType <= 7 {
			// DON'T SCAN THE URLS SCANNED BEFORE
			extension := huntExtensions(r.Request.URL.String(), fileType)
			if extension.Url != "" {
				FinalExtensions = append(FinalExtensions, extension)
			}
		}

		Finalresult = append(Finalresult, r.Request.URL.String())
	})

	// Start scraping on target
	c.Visit("http://" + target)
	c.Visit("https://" + target)
	c.Wait()
	return Finalresult, Finalsecrets, Finalendpoints, FinalExtensions
}

//huntSecrets
func huntSecrets(secretsFile []string, target string, body string) []scanner.Secret {
	secrets := SecretsMatch(body, secretsFile)
	return secrets
}

//SecretsMatch
func SecretsMatch(body string, secretsFile []string) []scanner.Secret {
	var secrets []scanner.Secret
	if len(secretsFile) == 0 {
		for _, secret := range scanner.GetRegexes() {
			if matched, err := regexp.Match(secret.Regex, []byte(body)); err == nil && matched {
				secrets = append(secrets, secret)
			}
		}
	} else {
		for _, secret := range secretsFile {
			if matched, err := regexp.Match(secret, []byte(body)); err == nil && matched {
				secrets = append(secrets, scanner.Secret{Name: "CustomFromFile", Description: "", Regex: secret, Poc: ""})
			}
		}
	}
	return secrets
}

//huntEndpoints
func huntEndpoints(endpointsFile []string, target string) []scanner.EndpointMatched {
	endpoints := EndpointsMatch(target, endpointsFile)
	return endpoints
}

//EndpointsMatch
func EndpointsMatch(target string, endpointsFile []string) []scanner.EndpointMatched {
	var endpoints []scanner.EndpointMatched
	matched := []string{}
	if len(endpointsFile) == 0 {
		for _, parameter := range scanner.GetJuicyParameters() {
			if strings.Contains(target, parameter) {
				matched = append(matched, parameter)
			}
			endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, Url: target})
		}
	} else {
		for _, parameter := range endpointsFile {
			if strings.Contains(target, parameter) {
				matched = append(matched, parameter)
			}
			endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, Url: target})
		}
	}
	return endpoints
}

//huntExtensions
func huntExtensions(target string, severity int) scanner.FileTypeMatched {
	var extension scanner.FileTypeMatched
	copyTarget := target
	for _, ext := range scanner.GetExtensions() {
		if ext.Severity <= severity {
			firstIndex := strings.Index(target, "?")
			if firstIndex > -1 {
				target = target[:firstIndex]
			}
			i := strings.LastIndex(target, ".")
			if i >= 0 && target[i:] == "."+ext.Extension {
				extension = scanner.FileTypeMatched{Filetype: ext, Url: copyTarget}
			}
		}
	}
	return extension
}

//RetrieveBody
func RetrieveBody(target string) string {
	sb, err := GetRequest(target)
	if err == nil && sb != "" {
		return sb
	}
	return ""
}

//isLinkOkay
func isLinkOkay(input string) bool {
	_, err := url.Parse(input)
	return err == nil
}
