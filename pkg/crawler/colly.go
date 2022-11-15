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
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"time"

	fileUtils "github.com/edoardottt/cariddi/internal/file"
	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
	urlUtils "github.com/edoardottt/cariddi/internal/url"
	"github.com/edoardottt/cariddi/pkg/input"
	"github.com/edoardottt/cariddi/pkg/output"
	"github.com/edoardottt/cariddi/pkg/scanner"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
)

// New it's the actual crawler engine.
// It controls all the behaviours of a scan
// (event handlers, secrets, errors, extensions and endpoints scanning).
func New(target string, txt string, html string, delayTime int, concurrency int,
	ignore string, ignoreTxt string, cache bool, timeout int, intensive bool, rua bool,
	proxy string, insecure bool, secretsFlag bool, secretsFile []string, plain bool, endpointsFlag bool,
	endpointsFile []string, fileType int, headers map[string]string, errorsFlag bool, infoFlag bool,
	debug bool, userAgent string) ([]string, []scanner.SecretMatched, []scanner.EndpointMatched,
	[]scanner.FileTypeMatched, []scanner.ErrorMatched, []scanner.InfoMatched) {
	// This is to avoid to insert into the crawler target regular
	// expression directories passed as input.
	var targetTemp, protocolTemp string

	// if there isn't a scheme use http.
	if !urlUtils.HasProtocol(target) {
		protocolTemp = "http"
		targetTemp = urlUtils.GetHost(protocolTemp + "://" + target)
	} else {
		protocolTemp = urlUtils.GetProtocol(target)
		targetTemp = urlUtils.GetHost(target)
	}

	if intensive {
		var err error
		targetTemp, err = urlUtils.GetRootHost(protocolTemp + "://" + targetTemp)

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	if targetTemp == "" {
		fmt.Println("The URL provided is not built in a proper way: " + target)
		os.Exit(1)
	}

	// clean target input
	target = urlUtils.RemoveProtocol(target)

	ignoreSlice := []string{}
	ignoreBool := false

	// if ignore -> produce the slice
	if ignore != "" {
		ignoreBool = true
		ignoreSlice = sliceUtils.CheckInputArray(ignore)
	}

	// if ignoreTxt -> produce the slice
	if ignoreTxt != "" {
		ignoreBool = true
		ignoreSlice = fileUtils.ReadFile(ignoreTxt)
	}

	FinalResults := []string{}
	FinalSecrets := []scanner.SecretMatched{}
	FinalEndpoints := []scanner.EndpointMatched{}
	FinalExtensions := []scanner.FileTypeMatched{}
	FinalErrors := []scanner.ErrorMatched{}
	FinalInfos := []scanner.InfoMatched{}

	// crawler creation
	c := CreateColly(delayTime, concurrency, cache, timeout, intensive, rua, proxy, insecure, userAgent, target)

	// On every request that Colly is making, print the URL it's currently visiting
	c.OnRequest(func(e *colly.Request) {
		fmt.Println(e.URL.String())
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(link) != 0 && link[0] != '#' {
			visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
		}
	})

	// On every script element which has src attribute call callback
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// On every link element which has href attribute call callback
	c.OnHTML("link[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// On every iframe element which has src attribute call callback
	c.OnHTML("iframe[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// On every svg element which has src attribute call callback
	c.OnHTML("svg[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// On every img element which has src attribute call callback
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// On every from element which has action attribute call callback
	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		link := e.Attr("action")
		visitHTMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//url", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//link", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//href", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//loc", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//fileurl", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, target, intensive, ignoreBool, debug, ignoreSlice, &FinalResults, e, c)
	})

	// Add headers (if needed) on each request
	if (len(headers)) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for header, value := range headers {
				r.Headers.Set(header, value)
			}
		})
	}

	c.OnResponse(func(r *colly.Response) {
		minBodyLentgh := 10

		lengthOk := len(string(r.Body)) > minBodyLentgh

		// if endpoints or secrets or filetype: scan
		if endpointsFlag || secretsFlag || (1 <= fileType && fileType <= 7) || errorsFlag || infoFlag {
			// HERE SCAN FOR SECRETS
			if secretsFlag && lengthOk {
				secretsSlice := huntSecrets(secretsFile, r.Request.URL.String(), string(r.Body))
				for _, elem := range secretsSlice {
					FinalSecrets = append(FinalSecrets, elem)
				}
			}
			// HERE SCAN FOR ENDPOINTS
			if endpointsFlag {
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
				if extension.URL != "" {
					FinalExtensions = append(FinalExtensions, extension)
				}
			}
			// HERE SCAN FOR ERRORS
			if errorsFlag {
				errorsSlice := huntErrors(r.Request.URL.String(), string(r.Body))
				for _, elem := range errorsSlice {
					FinalErrors = append(FinalErrors, elem)
				}
			}

			// HERE SCAN FOR INFOS
			if infoFlag {
				infosSlice := huntInfos(r.Request.URL.String(), string(r.Body))
				for _, elem := range infosSlice {
					FinalInfos = append(FinalInfos, elem)
				}
			}
		}
	})

	// Start scraping on target
	path, err := urlUtils.GetPath(protocolTemp + "://" + target)
	if err == nil {
		var (
			addPath     string
			absoluteURL string
		)

		if path == "" {
			addPath = "/"
		}

		absoluteURL = protocolTemp + "://" + target + addPath + "robots.txt"
		if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, ignoreSlice)) {
			err = c.Visit(absoluteURL)
			if err != nil && debug && !errors.Is(err, colly.ErrAlreadyVisited) {
				log.Println(err)
			}
		}

		absoluteURL = protocolTemp + "://" + target + addPath + "sitemap.xml"
		if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, ignoreSlice)) {
			err = c.Visit(absoluteURL)
			if err != nil && debug && !errors.Is(err, colly.ErrAlreadyVisited) {
				log.Println(err)
			}
		}
	}

	err = c.Visit(protocolTemp + "://" + target)
	if err != nil && debug && !errors.Is(err, colly.ErrAlreadyVisited) {
		log.Println(err)
	}

	// Setup graceful exit
	chanC := make(chan os.Signal, 1)
	lettersNum := 23
	cCount := 0

	signal.Notify(chanC, os.Interrupt)
	rand.Seed(time.Now().UnixNano())

	go func() {
		for range chanC {
			if cCount > 0 {
				os.Exit(1)
			}

			if !plain {
				fmt.Fprint(os.Stdout, "\r")
				fmt.Println("CTRL+C pressed: Exiting")
				cCount++
			}

			c.AllowedDomains = []string{sliceUtils.RandSeq(lettersNum)}
		}
	}()

	c.Wait()

	if html != "" {
		output.FooterHTML(html)
	}

	return FinalResults, FinalSecrets, FinalEndpoints, FinalExtensions, FinalErrors, FinalInfos
}

// CreateColly takes as input all the settings needed to instantiate
// a new Colly Collector object and it returns this object.
func CreateColly(delayTime int, concurrency int, cache bool, timeout int,
	intensive bool, rua bool, proxy string, insecure bool, userAgent string, target string) *colly.Collector {
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
	if timeout != input.TimeoutRequest {
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

	if userAgent != "" {
		c.UserAgent = userAgent
	}

	// Use a Proxy if needed
	if proxy != "" {
		err := c.SetProxy(proxy)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if insecure {
		c.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})
	}

	return c
}

// visitHTMLLink checks if the collector should visit a link or not.
func visitHTMLLink(link, protocolTemp, targetTemp, target string, intensive, ignoreBool, debug bool,
	ignoreSlice []string, finalResults *[]string, e *colly.HTMLElement, c *colly.Collector) {
	if len(link) != 0 {
		absoluteURL := urlUtils.AbsoluteURL(protocolTemp, targetTemp, e.Request.AbsoluteURL(link))
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		if (!intensive && urlUtils.SameDomain(protocolTemp+"://"+target, absoluteURL)) ||
			(intensive && intensiveOk(targetTemp, absoluteURL, debug)) {
			if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, ignoreSlice)) {
				err := c.Visit(absoluteURL)
				if !errors.Is(err, colly.ErrAlreadyVisited) {
					*finalResults = append(*finalResults, absoluteURL)

					if err != nil && debug {
						log.Println(err)
					}
				}
			}
		}
	}
}
