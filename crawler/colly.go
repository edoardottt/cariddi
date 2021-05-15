package crawler

import (
	"net/url"
	"regexp"
	"time"

	"github.com/edoardottt/cariddi/input"
	"github.com/edoardottt/cariddi/output"
	"github.com/edoardottt/cariddi/scanner"
	"github.com/gocolly/colly"
)

//Crawler
func Crawler(target string, delayTime int, concurrency int, secrets bool, secretsFile string, dataPost map[string]string) []string {

	//clean target input
	target = input.RemoveHeaders(target)

	var result []string
	// Instantiate  collector
	c := colly.NewCollector(
		colly.AllowedDomains(target),
		colly.Async(true),
		colly.URLFilters(
			regexp.MustCompile("(http://|https://|ftp://|)"+target+"*"),
		),
	)

	/*
		1. IF secrets flag enabled -> OK
			1.1. Declare regexes in the first part of the function.
			1.2. Print the entire page and see what actually is `page` string.
			1.3  Test with a custom website and see if actually it works properly.

			OR

			use the requests file to take the body...

			Let's see

			//scan for secrets here
			c.OnHTML("*", func(e *colly.HTMLElement) {
				page := e.Attr("html")
				for _, regex := range scanner.GetRegexes() {
					matched, err := regexp.MatchString(`a.b`, page)
					if err != nil {
						panic(err.Error())
					}
					if matched {
						print(regex.Name)
					}
				}
			})
	*/

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
		//result = append(result, e.Request.AbsoluteURL(link))
	})

	c.OnRequest(func(r *colly.Request) {
		//fmt.Println("Visiting", r.URL.String())
		// HERE SCAN FOR SECRETS
		if secrets {
			secretsSlice := huntSecrets(secretsFile, r.URL.String(), dataPost)
			for _, elem := range secretsSlice {
				output.EncapsulateCustomGreen(elem.Name, "Found in "+r.URL.String()+" "+elem.Regex+" matched!")
			}
		}
		result = append(result, r.URL.String())
	})

	// Start scraping on target
	c.Visit("https://" + target)
	c.Visit("http://" + target)
	c.Wait()
	return result
}

//huntSecrets
func huntSecrets(secretsFile string, target string, data map[string]string) []scanner.Secret {
	if secretsFile == "" {
		body := RetrieveBody(target, data)
		secrets := SecretsMatch(body)
		return secrets
	}
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
