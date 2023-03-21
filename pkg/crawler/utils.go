package crawler

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"

	urlUtils "github.com/edoardottt/cariddi/internal/url"
	"github.com/edoardottt/cariddi/pkg/scanner"
	"github.com/gocolly/colly"
)

type Event struct {
	ProtocolTemp string
	TargetTemp   string
	Target       string
	Intensive    bool
	Ignore       bool
	Debug        bool
	IgnoreSlice  []string
	URLs         *[]string
}

// visitHTMLLink checks if the collector should visit a link or not.
func visitHTMLLink(link string, event *Event, e *colly.HTMLElement, c *colly.Collector) {
	if len(link) != 0 && !strings.HasPrefix(link, "data:image") {
		absoluteURL := urlUtils.AbsoluteURL(event.ProtocolTemp, event.TargetTemp, e.Request.AbsoluteURL(link))
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		visitLink(event, c, absoluteURL)
	}
}

// visitXMLLink checks if the collector should visit a link or not.
func visitXMLLink(link string, event *Event, e *colly.XMLElement, c *colly.Collector) {
	if len(link) != 0 && !strings.HasPrefix(link, "data:image") {
		absoluteURL := urlUtils.AbsoluteURL(event.ProtocolTemp, event.TargetTemp, e.Request.AbsoluteURL(link))
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		visitLink(event, c, absoluteURL)
	}
}

// visitLink is a protocol agnostic wrapper to visit a link.
func visitLink(event *Event, c *colly.Collector, absoluteURL string) {
	if (!event.Intensive && urlUtils.SameDomain(event.ProtocolTemp+"://"+event.Target, absoluteURL)) ||
		(event.Intensive && intensiveOk(event.TargetTemp, absoluteURL, event.Debug)) {
		if !event.Ignore || (event.Ignore && !IgnoreMatch(absoluteURL, &event.IgnoreSlice)) {
			err := c.Visit(absoluteURL)
			if !errors.Is(err, colly.ErrAlreadyVisited) {
				*event.URLs = append(*event.URLs, absoluteURL)

				if err != nil && event.Debug {
					log.Println(err)
				}
			}
		}
	}
}

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
					secretFound := scanner.SecretMatched{Secret: secret, URL: url, Match: match[0]}
					secrets = append(secrets, secretFound)
				}
			}
		}
	} else {
		for _, secret := range *secretsFile {
			if matched, err := regexp.Match(secret, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(secret)
				match := re.FindStringSubmatch(body)
				secretScanned := scanner.Secret{Name: "CustomFromFile", Description: "", Regex: secret, Poc: ""}
				secretFound := scanner.SecretMatched{Secret: secretScanned, URL: url, Match: match[0]}
				secrets = append(secrets, secretFound)
			}
		}
	}

	return secrets
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
		for _, errorRegex := range errorItem.Regex {
			if matched, err := regexp.Match(errorRegex, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(errorRegex)
				match := re.FindStringSubmatch(body)
				errorFound := scanner.ErrorMatched{Error: errorItem, URL: url, Match: match[0]}
				errors = append(errors, errorFound)
			}
		}
	}

	return errors
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
		for _, infoRegex := range infoItem.Regex {
			if matched, err := regexp.Match(infoRegex, []byte(body)); err == nil && matched {
				re := regexp.MustCompile(infoRegex)
				match := re.FindStringSubmatch(body)
				infoFound := scanner.InfoMatched{Info: infoItem, URL: url, Match: match[0]}
				infos = append(infos, infoFound)
			}
		}
	}

	return infos
}

// RetrieveBody retrieves the body (in the response) of a url.
func RetrieveBody(target *string) string {
	sb, err := GetRequest(*target)
	if err == nil && sb != "" {
		return sb
	}

	return ""
}

// IgnoreMatch checks if the URL should be ignored or not.
func IgnoreMatch(url string, ignoreSlice *[]string) bool {
	for _, ignore := range *ignoreSlice {
		if strings.Contains(url, ignore) {
			return true
		}
	}

	return false
}

// intensiveOk checks if a given url can be crawled
// in intensive mode (if the 2nd level domain matches with
// the inputted target).
func intensiveOk(target string, urlInput string, debug bool) bool {
	root, err := urlUtils.GetRootHost(urlInput)
	if err != nil {
		if debug {
			fmt.Println(err.Error() + ": " + urlInput)
		}

		return false
	}

	return root == target
}
