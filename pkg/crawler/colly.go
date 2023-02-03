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

type Results struct {
	URLs       []string
	Secrets    []scanner.SecretMatched
	Endpoints  []scanner.EndpointMatched
	Extensions []scanner.FileTypeMatched
	Errors     []scanner.ErrorMatched
	Infos      []scanner.InfoMatched
}

type Scan struct {
	// Flags
	Cache         bool
	Debug         bool
	EndpointsFlag bool
	ErrorsFlag    bool
	InfoFlag      bool
	Intensive     bool
	Plain         bool
	Rua           bool
	SecretsFlag   bool
	Ignore        string
	IgnoreTxt     string
  JSON          bool
	HTML          string
	Proxy         string
	Target        string
	Txt           string
	UserAgent     string
	FileType      int
	Headers       map[string]string

	// Settings
	Concurrency int
	Delay       int
	Timeout     int

	// Storage
	SecretsSlice   []string
	EndpointsSlice []string
}

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
		targetTemp = urlUtils.GetHost(protocolTemp + "://" + scan.Target)
	} else {
		protocolTemp = urlUtils.GetProtocol(scan.Target)
		targetTemp = urlUtils.GetHost(scan.Target)
	}

	if scan.Intensive {
		var err error
		targetTemp, err = urlUtils.GetRootHost(protocolTemp + "://" + targetTemp)

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

	// On every request that Colly is making, print the URL it's currently visiting
	c.OnRequest(func(e *colly.Request) {
		if jsonl == false {
			fmt.Println(e.URL.String())
		}
	})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(link) != 0 && link[0] != '#' {
			visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
				ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
		}
	})

	// On every script element which has src attribute call callback
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// On every link element which has href attribute call callback
	c.OnHTML("link[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// On every iframe element which has src attribute call callback
	c.OnHTML("iframe[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// On every svg element which has src attribute call callback
	c.OnHTML("svg[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// On every img element which has src attribute call callback
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		link := e.Attr("src")
		visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// On every from element which has action attribute call callback
	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		link := e.Attr("action")
		visitHTMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//url", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//link", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//href", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//loc", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//fileurl", func(e *colly.XMLElement) {
		link := e.Text
		visitXMLLink(link, protocolTemp, targetTemp, scan.Target, scan.Intensive,
			ignoreBool, scan.Debug, ignoreSlice, &results.URLs, e, c)
	})

	// Add headers (if needed) on each request
	if (len(scan.Headers)) > 0 {
		c.OnRequest(func(r *colly.Request) {
			for header, value := range scan.Headers {
				r.Headers.Set(header, value)
			}
		})
	}

	c.OnResponse(func(r *colly.Response) {
		minBodyLentgh := 10
		lengthOk := len(string(r.Body)) > minBodyLentgh
		secrets := []scanner.SecretMatched{}
		parameters := []scanner.Parameter{}
		errors := []scanner.ErrorMatched{}
		infos := []scanner.InfoMatched{}
		filetype := scanner.FileType{}

		// if endpoints or secrets or filetype: scan
		if scan.EndpointsFlag || scan.SecretsFlag ||
			(1 <= scan.FileType && scan.FileType <= 7) || scan.ErrorsFlag || scan.InfoFlag {
			// HERE SCAN FOR SECRETS
			if scan.SecretsFlag && lengthOk {
				secretsSlice := huntSecrets(scan.SecretsSlice, r.Request.URL.String(), string(r.Body))
				results.Secrets = append(results.Secrets, secretsSlice...)
        secrets = append(secrets, secretsSlice...)
			}
			// HERE SCAN FOR ENDPOINTS
			if scan.EndpointsFlag {
				endpointsSlice := huntEndpoints(scan.EndpointsSlice, r.Request.URL.String())
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
          filetype = extension.Filetype
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
        infos = append(infos, infosSlice)
			}
		}
		if jsonl == true {
			output.GetJsonString(r, secrets, parameters, filetype, errors, infos)
		}
	})

	// Start scraping on target
	path, err := urlUtils.GetPath(protocolTemp + "://" + scan.Target)
	if err == nil {
		var (
			addPath     string
			absoluteURL string
		)

		if path == "" {
			addPath = "/"
		}

		if path == "" || path == "/" {
			absoluteURL = protocolTemp + "://" + scan.Target + addPath + "robots.txt"
			if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, ignoreSlice)) {
				err = c.Visit(absoluteURL)
				if err != nil && scan.Debug && !errors.Is(err, colly.ErrAlreadyVisited) {
					log.Println(err)
				}
			}

			absoluteURL = protocolTemp + "://" + scan.Target + addPath + "sitemap.xml"
			if !ignoreBool || (ignoreBool && !IgnoreMatch(absoluteURL, ignoreSlice)) {
				err = c.Visit(absoluteURL)
				if err != nil && scan.Debug && !errors.Is(err, colly.ErrAlreadyVisited) {
					log.Println(err)
				}
			}
		}
	}

	err = c.Visit(protocolTemp + "://" + scan.Target)
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
		err := c.SetProxy(proxy)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	c.WithTransport(&http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	})

	return c
}
