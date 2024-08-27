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
	"net/url"
	"os"
	"os/signal"
	"strings"
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
func New(scan *Scan) *Results {
	// This is to avoid to insert into the crawler target regular
	// expression directories passed as input.
	var targetTemp, protocolTemp string

	results := &Results{}

	// if there isn't a scheme use http.
	if !urlUtils.HasProtocol(scan.Target) {
		protocolTemp = "http"
		targetTemp = urlUtils.GetHost(fmt.Sprintf("%s://%s", protocolTemp, scan.Target))
	} else {
		protocolTemp = urlUtils.GetProtocol(scan.Target)
		targetTemp = urlUtils.GetHost(scan.Target)
	}

	if scan.Intensive {
		var err error
		targetTemp, err = urlUtils.GetRootHost(fmt.Sprintf("%s://%s", protocolTemp, targetTemp))

		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	if targetTemp == "" {
		fmt.Println("The URL provided is not built in a proper way: " + scan.Target)
		os.Exit(1)
	}

	// clean target input
	scan.Target = urlUtils.RemoveProtocol(scan.Target)

	ignoreSlice := []string{}
	ignoreBool := false

	// if ignore -> produce the slice
	if scan.Ignore != "" {
		ignoreBool = true
		ignoreSlice = sliceUtils.CheckInputArray(scan.Ignore)
	}

	// if ignoreTxt -> produce the slice
	if scan.IgnoreTxt != "" {
		ignoreBool = true
		ignoreSlice = fileUtils.ReadFile(scan.IgnoreTxt)
	}

	// crawler creation
	c := CreateColly(scan.Delay, scan.Concurrency, scan.Cache, scan.Timeout,
		scan.Intensive, scan.Rua, scan.Proxy, scan.UserAgent, scan.Target)

	event := &Event{
		ProtocolTemp: protocolTemp,
		TargetTemp:   targetTemp,
		Target:       scan.Target,
		Intensive:    scan.Intensive,
		Ignore:       ignoreBool,
		Debug:        scan.Debug,
		JSON:         scan.JSON,
		IgnoreSlice:  ignoreSlice,
		URLs:         &results.URLs,
	}

	registerHTMLEvents(c, event)
	registerXMLEvents(c, event)

	// Add headers (if needed) on each request
	if (len(scan.Headers)) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for header, value := range scan.Headers {
				r.Headers.Set(header, value)
			}
		})
	}

	c.OnResponse(func(r *colly.Response) {
		if !scan.JSON {
			fmt.Println(r.Request.URL)
		}

		var outputPath string

		if scan.StoreResp {
			var err error
			outputPath, err = output.StoreHTTPResponse(r, scan.OutputDir)

			if err != nil {
				log.Println(err)
			}
		}

		minBodyLentgh := 10
		lengthOk := len(string(r.Body)) > minBodyLentgh
		secrets := []scanner.SecretMatched{}
		parameters := []scanner.Parameter{}
		errors := []scanner.ErrorMatched{}
		infos := []scanner.InfoMatched{}
		filetype := &scanner.FileType{}

		// if endpoints or secrets or filetype: scan
		if scan.EndpointsFlag || scan.SecretsFlag ||
			(1 <= scan.FileType && scan.FileType <= 7) || scan.ErrorsFlag || scan.InfoFlag {
			// HERE SCAN FOR SECRETS
			if scan.SecretsFlag && lengthOk {
				secretsSlice := huntSecrets(r.Request.URL.String(), string(r.Body), &scan.SecretsSlice)
				results.Secrets = append(results.Secrets, secretsSlice...)
				secrets = append(secrets, secretsSlice...)
			}
			// HERE SCAN FOR ENDPOINTS
			if scan.EndpointsFlag {
				endpointsSlice := huntEndpoints(r.Request.URL.String(), &scan.EndpointsSlice)
				for _, elem := range endpointsSlice {
					if len(elem.Parameters) != 0 {
						results.Endpoints = append(results.Endpoints, elem)
						parameters = append(parameters, elem.Parameters...)
					}
				}
			}
			// HERE SCAN FOR EXTENSIONS
			if 1 <= scan.FileType && scan.FileType <= 7 {
				extension := huntExtensions(r.Request.URL.String(), scan.FileType)
				if extension.URL != "" {
					results.Extensions = append(results.Extensions, extension)
					filetype = &extension.Filetype
				}
			}
			// HERE SCAN FOR ERRORS
			if scan.ErrorsFlag {
				errorsSlice := huntErrors(r.Request.URL.String(), string(r.Body))
				results.Errors = append(results.Errors, errorsSlice...)
				errors = append(errors, errorsSlice...)
			}

			// HERE SCAN FOR INFOS
			if scan.InfoFlag {
				infosSlice := huntInfos(r.Request.URL.String(), string(r.Body))
				results.Infos = append(results.Infos, infosSlice...)
				infos = append(infos, infosSlice...)
			}
		}

		if scan.JSON {
			jsonOutput, err := output.GetJSONString(
				r, secrets, parameters, filetype, errors, infos, outputPath,
			)

			if err == nil {
				fmt.Println(string(jsonOutput))
			} else {
				log.Println(err)
			}
		}
	})

	// Start scraping on target
	path, err := urlUtils.GetPath(fmt.Sprintf("%s://%s", protocolTemp, scan.Target))
	if err == nil {
		var (
			addPath     string
			absoluteURL string
		)

		if path == "" {
			addPath = "/"
		}

		if path == "" || path == "/" {
			absoluteURL = fmt.Sprintf("%s://%s%srobots.txt", protocolTemp, scan.Target, addPath)
			if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, &ignoreSlice)) {
				err = c.Visit(absoluteURL)
				if err != nil && scan.Debug && !errors.Is(err, colly.ErrAlreadyVisited) {
					log.Println(err)
				}
			}

			absoluteURL = protocolTemp + "://" + scan.Target + addPath + "sitemap.xml"
			if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, &ignoreSlice)) {
				err = c.Visit(absoluteURL)
				if err != nil && scan.Debug && !errors.Is(err, colly.ErrAlreadyVisited) {
					log.Println(err)
				}
			}
		}
	}

	err = c.Visit(fmt.Sprintf("%s://%s", protocolTemp, scan.Target))
	if err != nil && scan.Debug && !errors.Is(err, colly.ErrAlreadyVisited) {
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

			if !scan.Plain {
				fmt.Fprint(os.Stdout, "\r")
				fmt.Println("CTRL+C pressed: Exiting")

				cCount++
			}

			c.AllowedDomains = []string{sliceUtils.RandSeq(lettersNum)}
		}
	}()

	c.Wait()

	if scan.HTML != "" {
		output.FooterHTML(scan.HTML)
	}

	return results
}

// CreateColly takes as input all the settings needed to instantiate
// a new Colly Collector object and it returns this object.
func CreateColly(delayTime int, concurrency int, cache bool, timeout int,
	intensive bool, rua bool, proxy string, userAgent string, target string) *colly.Collector {
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
		proxyParsed, err := url.Parse(proxy)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		c.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true},
			Proxy:             http.ProxyURL(proxyParsed),
			DisableKeepAlives: true,
		})
	} else {
		c.WithTransport(&http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		})
	}

	return c
}

// registerHTMLEvents registers the associated functions for each
// HTML event triggering an action.
func registerHTMLEvents(c *colly.Collector, event *Event) {
	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(link) != 0 && link[0] != '#' {
			visitHTMLLink(link, event, e, c)
		}
	})

	// On every script element which has src attribute call callback
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("src"), event, e, c)
	})

	// On every link element which has href attribute call callback
	c.OnHTML("link[href]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("href"), event, e, c)
	})

	// On every iframe element which has src attribute call callback
	c.OnHTML("iframe[src]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("src"), event, e, c)
	})

	// On every svg element which has src attribute call callback
	c.OnHTML("svg[src]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("src"), event, e, c)
	})

	// On every img element which has src attribute call callback
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("src"), event, e, c)
	})

	// On every from element which has action attribute call callback
	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("action"), event, e, c)
	})
}

// registerXMLEvents registers the associated functions for each
// XML event triggering an action.
func registerXMLEvents(c *colly.Collector, event *Event) {
	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//url", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//link", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//href", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//loc", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//fileurl", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e, c)
	})
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
