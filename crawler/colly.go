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
func Crawler(target string, delayTime int, concurrency int, secrets bool, secretsFile string, plain bool, endpoints bool, endpointsFile string) ([]string, []scanner.SecretMatched, []scanner.EndpointMatched) {

	//clean target input
	target = input.RemoveHeaders(target)

	var Finalresult []string
	var Finalsecrets []scanner.SecretMatched
	var Finalendpoints []scanner.EndpointMatched

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

	c.OnRequest(func(r *colly.Request) {
		// HERE SCAN FOR SECRETS
		if secrets {
			// DON'T SCAN THE URLS SCANNED BEFORE
			secretsSlice := huntSecrets(secretsFile, r.URL.String())
			for _, elem := range secretsSlice {

				secretFound := scanner.SecretMatched{Secret: elem, Url: r.URL.String()}
				Finalsecrets = append(Finalsecrets, secretFound)
			}
		}
		// HERE SCAN FOR ENDPOINTS
		if endpoints {
			// DON'T SCAN THE URLS SCANNED BEFORE
			endpointsSlice := huntEndpoints(endpointsFile, r.URL.String())
			for _, elem := range endpointsSlice {
				if len(elem.Parameters) != 0 {
					Finalendpoints = append(Finalendpoints, elem)
				}
			}
		}
		Finalresult = append(Finalresult, r.URL.String())
	})

	// Start scraping on target
	c.Visit("http://" + target)
	c.Visit("https://" + target)
	c.Wait()
	return Finalresult, Finalsecrets, Finalendpoints
}

//huntSecrets
func huntSecrets(secretsFile string, target string) []scanner.Secret {
	if secretsFile == "" {
		body := RetrieveBody(target)
		secrets := SecretsMatch(body)
		return secrets
	}

	// HERE ---> ELSE SECRETS FILE !

	return scanner.GetRegexes()
}

//SecretsMatch
func SecretsMatch(body string) []scanner.Secret {
	var secrets []scanner.Secret
	for _, secret := range scanner.GetRegexes() {
		if matched, err := regexp.Match(secret.Regex, []byte(body)); err == nil && matched {
			secrets = append(secrets, secret)
		}
	}
	return secrets
}

//huntEndpoints
func huntEndpoints(endpointsFile string, target string) []scanner.EndpointMatched {
	if endpointsFile == "" {
		endpoints := EndpointsMatch(target)
		return endpoints
	}

	// HERE ---> ELSE SECRETS FILE !

	return nil
}

//EndpointsMatch
func EndpointsMatch(target string) []scanner.EndpointMatched {
	var endpoints []scanner.EndpointMatched
	matched := []string{}
	for _, parameter := range scanner.GetJuicyParameters() {
		if strings.Contains(target, parameter) {
			matched = append(matched, parameter)
		}
		endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, Url: target})
	}
	return endpoints
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
