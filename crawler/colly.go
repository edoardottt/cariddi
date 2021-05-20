package crawler

import (
	"net/url"
	"regexp"
	"time"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/gocolly/colly"
)

//Crawler
func Crawler(target string, delayTime int, concurrency int, secrets bool, secretsFile string, plain bool, dataPost map[string]string) ([]string, []scanner.SecretMatched) {

	//clean target input
	target = input.RemoveHeaders(target)

	var Finalresult []string
	var Finalsecrets []scanner.SecretMatched
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
			secretsSlice := huntSecrets(secretsFile, r.URL.String(), dataPost)
			for _, elem := range secretsSlice {

				secretFound := scanner.SecretMatched{Secret: elem, Url: r.URL.String()}
				Finalsecrets = append(Finalsecrets, secretFound)
			}
		}
		Finalresult = append(Finalresult, r.URL.String())
	})

	// Start scraping on target
	c.Visit("http://" + target)
	c.Visit("https://" + target)
	c.Wait()
	return Finalresult, Finalsecrets
}

//huntSecrets
func huntSecrets(secretsFile string, target string, data map[string]string) []scanner.Secret {
	if secretsFile == "" {
		body := RetrieveBody(target, data)
		secrets := SecretsMatch(body)
		return secrets
	}

	// HERE ---> ELSE SECRETS FILE !

	return scanner.GetRegexes()
}

//RetrieveBody
func RetrieveBody(target string, data map[string]string) string {
	sb, err := GetRequest(target)
	if err == nil && sb != "" {
		return sb
	}
	sb, err = PostRequest(target, data)
	if err == nil && sb != "" {
		return sb
	}
	return ""
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

//isLinkOkay
func isLinkOkay(input string) bool {
	_, err := url.Parse(input)
	return err == nil
}
