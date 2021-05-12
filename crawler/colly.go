package crawler

import (
	"fmt"

	"github.com/edoardottt/cariddi/input"
	"github.com/gocolly/colly"
)

//Crawler
func Crawler(target string) []string {

	//clean target input
	target = input.RemoveHeaders(target)

	var result []string
	// Instantiate  collector
	c := colly.NewCollector(
		colly.AllowedDomains(target),
		colly.Async(true),
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

	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})
	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
		result = append(result, e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	// THEN AFTER TESTS COMMENT THIS.
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on target
	c.Visit("http://" + target)
	c.Visit("https://" + target)
	c.Wait()
	return result
}
