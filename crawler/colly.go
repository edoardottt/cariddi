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

package crawler

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/edoardottt/cariddi/output"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/edoardottt/cariddi/utils"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

//Crawler it's the actual crawler engine.
//It controls all the behaviours of a scan
//(event handlers, secrets, errors, extensions and endpoints scanning)
func Crawler(target string, txt string, html string, delayTime int, concurrency int,
	ignore string, ignoreTxt string, cache bool, timeout int, intensive bool, rua bool,
	proxy string, secrets bool, secretsFile []string, plain bool, endpoints bool,
	endpointsFile []string, fileType int, headers map[string]string,
	errors bool, info bool, debug bool) ([]string, []scanner.SecretMatched, []scanner.EndpointMatched,
	[]scanner.FileTypeMatched, []scanner.ErrorMatched, []scanner.InfoMatched) {

	// This is to avoid to insert into the crawler target regular
	// expression directories passed as input.
	var targetTemp string
	var protocolTemp string

	// if there isn't a scheme use http.
	if !utils.HasProtocol(target) {
		protocolTemp = "http"
		targetTemp = utils.GetHost(protocolTemp + "://" + target)
	} else {
		protocolTemp = utils.GetProtocol(target)
		targetTemp = utils.GetHost(target)
	}

	if intensive {
		var err error
		targetTemp, err = utils.GetRootHost(protocolTemp + "://" + targetTemp)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	if targetTemp == "" {
		fmt.Println("The URL provided is not built in a proper way: " + target)
		os.Exit(1)
	}

	//clean target input
	target = utils.RemoveProtocol(target)

	var ignoreSlice []string
	ignoreBool := false
	//if ignore -> produce the slice
	if ignore != "" {
		ignoreBool = true
		ignoreSlice = utils.CheckInputArray(ignore)
	}

	//if ignoreTxt -> produce the slice
	if ignoreTxt != "" {
		ignoreBool = true
		ignoreSlice = utils.ReadFile(ignoreTxt)
	}

	var FinalResults []string
	var FinalSecrets []scanner.SecretMatched
	var FinalEndpoints []scanner.EndpointMatched
	var FinalExtensions []scanner.FileTypeMatched
	var FinalErrors []scanner.ErrorMatched
	var FinalInfos []scanner.InfoMatched

	//crawler creation
	c := CreateColly(delayTime, concurrency, cache, timeout, intensive, rua, proxy, target)

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(link) != 0 && link[0] != '#' {
			absoluteUrl := utils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			if (!intensive && utils.SameDomain(protocolTemp+"://"+target, absoluteUrl)) ||
				(intensive && intensiveOk(targetTemp, absoluteUrl)) {
				if ignoreBool {
					if !IgnoreMatch(link, ignoreSlice) {
						FinalResults = append(FinalResults, absoluteUrl)
						err := c.Visit(absoluteUrl)
						if err != nil && debug && err != colly.ErrAlreadyVisited {
							log.Println(err)
						}
					}
				} else {
					FinalResults = append(FinalResults, absoluteUrl)
					err := c.Visit(absoluteUrl)
					if err != nil && debug && err != colly.ErrAlreadyVisited {
						log.Println(err)
					}
				}
			}
		}
	})

	// On every script element which has src attribute call callback
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		if len(link) != 0 {
			absoluteUrl := utils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			if (!intensive && utils.SameDomain(protocolTemp+"://"+target, absoluteUrl)) ||
				(intensive && intensiveOk(targetTemp, absoluteUrl)) {
				if ignoreBool {
					if !IgnoreMatch(link, ignoreSlice) {
						FinalResults = append(FinalResults, absoluteUrl)
						err := c.Visit(absoluteUrl)
						if err != nil && debug && err != colly.ErrAlreadyVisited {
							log.Println(err)
						}
					}
				} else {
					FinalResults = append(FinalResults, absoluteUrl)
					err := c.Visit(absoluteUrl)
					if err != nil && debug && err != colly.ErrAlreadyVisited {
						log.Println(err)
					}
				}
			}
		}
	})

	// On every link element which has href attribute call callback
	c.OnHTML("link[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(link) != 0 {
			absoluteUrl := utils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			if (!intensive && utils.SameDomain(protocolTemp+"://"+target, absoluteUrl)) ||
				(intensive && intensiveOk(targetTemp, absoluteUrl)) {
				if ignoreBool {
					if !IgnoreMatch(link, ignoreSlice) {
						FinalResults = append(FinalResults, absoluteUrl)
						err := c.Visit(absoluteUrl)
						if err != nil && debug && err != colly.ErrAlreadyVisited {
							log.Println(err)
						}
					}
				} else {
					FinalResults = append(FinalResults, absoluteUrl)
					err := c.Visit(absoluteUrl)
					if err != nil && debug && err != colly.ErrAlreadyVisited {
						log.Println(err)
					}
				}
			}
		}
	})

	// On every iframe element which has src attribute call callback
	c.OnHTML("iframe[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		if len(link) != 0 {
			absoluteUrl := utils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			if (!intensive && utils.SameDomain(protocolTemp+"://"+target, absoluteUrl)) ||
				(intensive && intensiveOk(targetTemp, absoluteUrl)) {
				if ignoreBool {
					if !IgnoreMatch(link, ignoreSlice) {
						FinalResults = append(FinalResults, absoluteUrl)
						err := c.Visit(absoluteUrl)
						if err != nil && debug && err != colly.ErrAlreadyVisited {
							log.Println(err)
						}
					}
				} else {
					FinalResults = append(FinalResults, absoluteUrl)
					err := c.Visit(absoluteUrl)
					if err != nil && debug && err != colly.ErrAlreadyVisited {
						log.Println(err)
					}
				}
			}
		}
	})

	// On every from element which has action attribute call callback
	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		link := e.Attr("action")
		if len(link) != 0 {
			absoluteUrl := utils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			if (!intensive && utils.SameDomain(protocolTemp+"://"+target, absoluteUrl)) ||
				(intensive && intensiveOk(targetTemp, absoluteUrl)) {
				if ignoreBool {
					if !IgnoreMatch(link, ignoreSlice) {
						FinalResults = append(FinalResults, absoluteUrl)
						err := c.Visit(absoluteUrl)
						if err != nil && debug && err != colly.ErrAlreadyVisited {
							log.Println(err)
						}
					}
				} else {
					FinalResults = append(FinalResults, absoluteUrl)
					err := c.Visit(absoluteUrl)
					if err != nil && debug && err != colly.ErrAlreadyVisited {
						log.Println(err)
					}
				}
			}
		}
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//urlset/url/loc", func(e *colly.XMLElement) {
		link := e.Text
		if len(link) != 0 {
			absoluteUrl := utils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
			// Visit link found on page
			// Only those links are visited which are in AllowedDomains
			if (!intensive && utils.SameDomain(protocolTemp+"://"+target, absoluteUrl)) ||
				(intensive && intensiveOk(targetTemp, absoluteUrl)) {
				if ignoreBool {
					if !IgnoreMatch(link, ignoreSlice) {
						FinalResults = append(FinalResults, absoluteUrl)
						err := c.Visit(absoluteUrl)
						if err != nil && debug && err != colly.ErrAlreadyVisited {
							log.Println(err)
						}
					}
				} else {
					FinalResults = append(FinalResults, absoluteUrl)
					err := c.Visit(absoluteUrl)
					if err != nil && debug && err != colly.ErrAlreadyVisited {
						log.Println(err)
					}
				}
			}
		}
	})

	//Add headers (if needed) on each request
	if (len(headers)) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for header, value := range headers {
				r.Headers.Set(header, value)
			}
		})
	}

	c.OnResponse(func(r *colly.Response) {

		fmt.Println(r.Request.URL.String())

		lengthOk := len(string(r.Body)) > 10

		//if endpoints or secrets or filetype: scan
		if endpoints || secrets || (1 <= fileType && fileType <= 7) || errors || info {
			// HERE SCAN FOR SECRETS
			if secrets && lengthOk {
				secretsSlice := huntSecrets(secretsFile, r.Request.URL.String(), string(r.Body))
				FinalSecrets = append(FinalSecrets, secretsSlice...)
			}
			// HERE SCAN FOR ENDPOINTS
			if endpoints {
				endpointsSlice := huntEndpoints(endpointsFile, r.Request.URL.String())
				for _, elem := range endpointsSlice {
					if len(elem.Parameters) != 0 {
						FinalEndpoints = append(FinalEndpoints, elem)
					}
				}
			}
			// HERE SCAN FOR EXTENSIONS
			if 1 <= fileType && fileType <= 7 {
				extension := huntExtensions(r.Request.URL.String(), fileType)
				if extension.Url != "" {
					FinalExtensions = append(FinalExtensions, extension)
				}
			}
			// HERE SCAN FOR ERRORS
			if errors {
				errorsSlice := huntErrors(r.Request.URL.String(), string(r.Body))
				FinalErrors = append(FinalErrors, errorsSlice...)
			}

			// HERE SCAN FOR INFOS
			if info {
				infosSlice := huntInfos(r.Request.URL.String(), string(r.Body))
				FinalInfos = append(FinalInfos, infosSlice...)
			}
		}
	})

	// Start scraping on target
	path, err := utils.GetPath(protocolTemp + "://" + target)
	if err == nil {
		if path == "" {
			err = c.Visit(protocolTemp + "://" + target + "/" + "robots.txt")
			if err != nil && debug && err != colly.ErrAlreadyVisited {
				log.Println(err)
			}
			err = c.Visit(protocolTemp + "://" + target + "/" + "sitemap.xml")
			if err != nil && debug && err != colly.ErrAlreadyVisited {
				log.Println(err)
			}
		} else if path == "/" {
			err = c.Visit(protocolTemp + "://" + target + "robots.txt")
			if err != nil && debug && err != colly.ErrAlreadyVisited {
				log.Println(err)
			}
			err = c.Visit(protocolTemp + "://" + target + "sitemap.xml")
			if err != nil && debug && err != colly.ErrAlreadyVisited {
				log.Println(err)
			}
		}
	}
	err = c.Visit(protocolTemp + "://" + target)
	if err != nil && debug && err != colly.ErrAlreadyVisited {
		log.Println(err)
	}
	c.Wait()
	if html != "" {
		output.FooterHTML(html)
	}
	return FinalResults, FinalSecrets, FinalEndpoints, FinalExtensions, FinalErrors, FinalInfos
}

//CreateColly takes as input all the settings needed to instantiate
//a new Colly Collector object and it returns this object.
func CreateColly(delayTime int, concurrency int, cache bool, timeout int,
	intensive bool, rua bool, proxy string, target string) *colly.Collector {

	c := colly.NewCollector(
		colly.Async(true),
	)
	c.IgnoreRobotsTxt = true
	c.AllowURLRevisit = false

	err := c.Limit(
		&colly.LimitRule{
			Parallelism: concurrency,
			Delay:       time.Duration(delayTime) * time.Second,
			DomainGlob:  "*" + target,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Using timeout if needed
	if timeout != 10 {
		c.SetRequestTimeout(time.Second * time.Duration(timeout))
	}

	// Using cache if needed
	if cache {
		c.CacheDir = ".cariddi_cache"
	}

	// Use a Random User Agent for each request if needed
	if rua {
		extensions.RandomUserAgent(c)
	} else {
		// Avoid using the default colly user agent
		c.UserAgent = GenerateRandomUserAgent()
	}

	// Use a Proxy if needed
	if proxy != "" {
		err := c.SetProxy(proxy)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	return c
}

//huntSecrets hunts for secrets
func huntSecrets(secretsFile []string, target string, body string) []scanner.SecretMatched {
	secrets := SecretsMatch(target, body, secretsFile)
	return secrets
}

//SecretsMatch checks if a body matches some secrets
func SecretsMatch(url string, body string, secretsFile []string) []scanner.SecretMatched {
	var secrets []scanner.SecretMatched
	if len(secretsFile) == 0 {
		for _, secret := range scanner.GetSecretRegexes() {
			if matched, err := regexp.Match(secret.Regex, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(secret.Regex)
				match := re.FindStringSubmatch(body)
				// Avoiding false positives
				var isFalsePositive = false
				for _, falsePositive := range secret.FalsePositives {
					if strings.Contains(strings.ToLower(match[0]), falsePositive) {
						isFalsePositive = true
						break
					}
				}
				if !isFalsePositive {
					secretFound := scanner.SecretMatched{Secret: secret, Url: url, Match: match[0]}
					secrets = append(secrets, secretFound)
				}
			}
		}
	} else {
		for _, secret := range secretsFile {
			if matched, err := regexp.Match(secret, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(secret)
				match := re.FindStringSubmatch(body)
				secretScanned := scanner.Secret{Name: "CustomFromFile", Description: "", Regex: secret, Poc: ""}
				secretFound := scanner.SecretMatched{Secret: secretScanned, Url: url, Match: match[0]}
				secrets = append(secrets, secretFound)
			}
		}
	}
	return secrets
}

//huntEndpoints hunts for juicy endpoints
func huntEndpoints(endpointsFile []string, target string) []scanner.EndpointMatched {
	endpoints := EndpointsMatch(target, endpointsFile)
	return endpoints
}

//EndpointsMatch check if an endpoint matches a juicy parameter
func EndpointsMatch(target string, endpointsFile []string) []scanner.EndpointMatched {
	var endpoints []scanner.EndpointMatched
	matched := []scanner.Parameter{}
	parameters := utils.RetrieveParameters(target)
	if len(endpointsFile) == 0 {
		for _, parameter := range scanner.GetJuicyParameters() {
			for _, param := range parameters {
				if strings.ToLower(param) == parameter.Parameter {
					matched = append(matched, parameter)
				}
				endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, Url: target})
			}
		}
	} else {

		for _, parameter := range endpointsFile {
			for _, param := range parameters {
				if param == parameter {
					matched = append(matched, scanner.Parameter{Parameter: parameter, Attacks: []string{}})
				}
				endpoints = append(endpoints, scanner.EndpointMatched{Parameters: matched, Url: target})
			}
		}
	}
	return endpoints
}

//huntExtensions hunts for extensions
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
			if i >= 0 && strings.ToLower(target[i:]) == "."+ext.Extension {
				extension = scanner.FileTypeMatched{Filetype: ext, Url: copyTarget}
			}
		}
	}
	return extension
}

//huntErrors hunts for errors
func huntErrors(target string, body string) []scanner.ErrorMatched {
	errorsSlice := ErrorsMatch(target, body)
	return errorsSlice
}

//ErrorsMatch checks the patterns for errors
func ErrorsMatch(url string, body string) []scanner.ErrorMatched {
	var errors []scanner.ErrorMatched
	for _, errorItem := range scanner.GetErrorRegexes() {
		for _, errorRegex := range errorItem.Regex {
			if matched, err := regexp.Match(errorRegex, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(errorRegex)
				match := re.FindStringSubmatch(body)
				errorFound := scanner.ErrorMatched{Error: errorItem, Url: url, Match: match[0]}
				errors = append(errors, errorFound)
			}
		}
	}
	return errors
}

//huntInfos hunts for infos
func huntInfos(target string, body string) []scanner.InfoMatched {
	infosSlice := InfoMatch(target, body)
	return infosSlice
}

//InfoMatch checks the patterns for infos
func InfoMatch(url string, body string) []scanner.InfoMatched {
	var infos []scanner.InfoMatched
	for _, infoItem := range scanner.GetInfoRegexes() {
		for _, infoRegex := range infoItem.Regex {
			if matched, err := regexp.Match(infoRegex, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(infoRegex)
				match := re.FindStringSubmatch(body)
				infoFound := scanner.InfoMatched{Info: infoItem, Url: url, Match: match[0]}
				infos = append(infos, infoFound)
			}
		}
	}
	return infos
}

//RetrieveBody retrieves the body (in the response) of a url
func RetrieveBody(target string) string {
	sb, err := GetRequest(target)
	if err == nil && sb != "" {
		return sb
	}
	return ""
}

//IgnoreMatch checks if the URL should be ignored or not.
func IgnoreMatch(url string, ignoreSlice []string) bool {
	for _, ignore := range ignoreSlice {
		if strings.Contains(url, ignore) {
			return true
		}
	}
	return false
}

//intensiveOk checks if a given url can be crawled
//in intensive mode (if the 2nd level domain matches with
//the inputted target).
func intensiveOk(target string, urlInput string) bool {
	root, err := utils.GetRootHost(urlInput)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	return root == target
}
