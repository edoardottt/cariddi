package crawler

import (
	"fmt"

	"github.com/gocolly/colly"
)

//Crawler
func Crawler(target string) []string {
	var result []string
	// Instantiate  collector
	c := colly.NewCollector(
		colly.AllowedDomains(target),
		colly.Async(true),
	)

	c.OnHTML("*", func(e *colly.HTMLElement) {

	})

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
	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	// Start scraping on target
	c.Visit("http://" + target)
	c.Visit("https://" + target)
	c.Wait()
	return result
}
