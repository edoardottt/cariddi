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

	@Author:      edoardottt, https://edoardottt.com

	@License: https://github.com/edoardottt/cariddi/blob/main/LICENSE

*/

package crawler

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	fileUtils "github.com/edoardottt/cariddi/internal/file"
	sliceUtils "github.com/edoardottt/cariddi/internal/slice"
	urlUtils "github.com/edoardottt/cariddi/internal/url"
	"github.com/edoardottt/cariddi/pkg/input"
	"github.com/edoardottt/cariddi/pkg/output"
	"github.com/edoardottt/cariddi/pkg/scanner"
	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
)

// New it's the actual crawler engine.
// It controls all the behaviours of a scan
// (event handlers, secrets, errors, extensions and endpoints scanning).
// When ctx is canceled the crawl stops and the results found so far
// are returned.
func New(ctx context.Context, scan *Scan) *Results {
	results := &Results{}

	// Colly callbacks run concurrently: resultsMu guards results.
	var resultsMu sync.Mutex

	// if there isn't a scheme use http.
	target := scan.Target
	if !urlUtils.HasScheme(target) {
		target = "http://" + target
	}

	protocolTemp := urlUtils.GetProtocol(target)
	// Hosts are case insensitive (unlike paths).
	targetTemp := strings.ToLower(urlUtils.GetHost(target))

	if targetTemp == "" {
		// Skip the target: the other targets can still be crawled.
		log.Println("The URL provided is not built in a proper way: " + scan.Target)

		return results
	}

	// The scope tells which URLs can be crawled: only the target host and
	// port or, in intensive mode, its root domain and all the subdomains.
	scope, err := urlUtils.NewScope(fmt.Sprintf("%s://%s", protocolTemp, targetTemp), scan.Intensive)
	if err != nil {
		// Skip the target (e.g. a TLD in intensive mode).
		log.Println(err.Error() + ": " + scan.Target)

		return results
	}

	ignoreSlice := []string{}
	ignoreBool := false

	// if ignore -> produce the slice
	if scan.Ignore != "" {
		ignoreBool = true
		ignoreSlice = sliceUtils.CheckInputArray(scan.Ignore)
	}

	// if ignoreTxt -> produce the slice (trimmed as for -i: the
	// spaces around a line would never match a URL)
	if scan.IgnoreTxt != "" {
		ignoreBool = true

		for _, line := range fileUtils.ReadFile(scan.IgnoreTxt) {
			ignoreSlice = append(ignoreSlice, strings.TrimSpace(line))
		}
	}

	// Build the secret regexes only once per target.
	var secretScanner *scanner.SecretScanner
	if scan.SecretsFlag {
		secretScanner = newSecretScanner(scan.SecretsSlice)
	}

	// The connections kept for reuse serve only the crawl of this target.
	transport := newTransport(scan.Concurrency, scan.Proxy)
	defer transport.CloseIdleConnections()

	// crawler creation
	c := CreateColly(scan.Delay, scan.Concurrency, scan.Timeout,
		scan.Cache, scan.Rua, transport, scan.UserAgent)

	// When ctx is canceled (CTRL+C) the requests in flight and the queued
	// ones fail at once, so the crawl ends with the results found so far.
	c.Context = ctx

	// wait waits for the queued requests, but returns at once on CTRL+C:
	// the requests left fail, but each one still waits for the delay.
	wait := func() {
		done := make(chan struct{})

		go func() {
			c.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-ctx.Done():
		}
	}

	event := &Event{
		ProtocolTemp: protocolTemp,
		TargetTemp:   targetTemp,
		Scope:        scope,
		Ignore:       ignoreBool,
		Debug:        scan.Debug,
		JSON:         scan.JSON,
		IgnoreSlice:  ignoreSlice,
		URLs:         &results.URLs,
		MaxDepth:     scan.MaxDepth,
	}

	registerHTMLEvents(c, event)
	registerXMLEvents(c, event)
	c.SetRedirectHandler(redirectHandler(event))

	// addURL adds a crawled URL to the results and prints it.
	addURL := func(u string) {
		resultsMu.Lock()

		results.URLs = append(results.URLs, u)

		resultsMu.Unlock()

		if !scan.JSON {
			fmt.Println(u)
		}
	}

	c.OnRequest(func(r *colly.Request) {
		// Add headers (if needed) on each request
		if (len(scan.Headers)) > 0 {
			for header, value := range scan.Headers {
				r.Headers.Set(header, value)
			}
		}

		// The sitemap probe is added only if it's a valid sitemap (OnResponse).
		if isSitemapProbe(r) {
			return
		}

		addURL(r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		// Ignore the sitemap probe if it isn't a valid sitemap (as a 404).
		if isSitemapProbe(r.Request) {
			if !isSitemap(r.Body) {
				return
			}

			addURL(r.Request.URL.String())
		}

		if scan.StoreResp {
			err := output.StoreHTTPResponse(r)
			if err != nil {
				log.Println(err)
			}
		}

		bodyStr := string(r.Body)

		minBodyLentgh := 10
		lengthOk := len(bodyStr) > minBodyLentgh
		secrets := []scanner.SecretMatched{}
		endpoints := []scanner.EndpointMatched{}
		parameters := []scanner.Parameter{}
		extensionsFound := []scanner.FileTypeMatched{}
		errors := []scanner.ErrorMatched{}
		infos := []scanner.InfoMatched{}
		filetype := &scanner.FileType{}

		// Skip if no scanning is enabled
		if !scan.EndpointsFlag && !scan.SecretsFlag && (scan.FileType < 1 || scan.FileType > 7) &&
			!scan.ErrorsFlag && !scan.InfoFlag && !scan.JSON {
			return
		}

		// HERE SCAN FOR SECRETS
		if scan.SecretsFlag && lengthOk &&
			!sliceUtils.Contains(scan.IgnoreExtensions, urlUtils.GetURLExtension(r.Request.URL)) {
			secretsSlice := huntSecrets(r.Request.URL.String(), bodyStr, secretScanner)
			secrets = append(secrets, secretsSlice...)
		}
		// The URL of an error page (e.g. a 404) is not a finding,
		// so endpoints and extensions are hunted in 2xx responses only.
		// HERE SCAN FOR ENDPOINTS
		if scan.EndpointsFlag && isSuccess(r) {
			endpointsSlice := huntEndpoints(r.Request.URL.String(), &scan.EndpointsSlice)
			for _, elem := range endpointsSlice {
				if len(elem.Parameters) != 0 {
					endpoints = append(endpoints, elem)
					parameters = append(parameters, elem.Parameters...)
				}
			}
		}
		// HERE SCAN FOR EXTENSIONS
		if 1 <= scan.FileType && scan.FileType <= 7 && isSuccess(r) {
			extension := huntExtensions(r.Request.URL.String(), scan.FileType)
			if extension.URL != "" {
				extensionsFound = append(extensionsFound, extension)
				filetype = &extension.Filetype
			}
		}
		// HERE SCAN FOR ERRORS
		if scan.ErrorsFlag && !sliceUtils.Contains(scan.IgnoreExtensions, urlUtils.GetURLExtension(r.Request.URL)) {
			errorsSlice := huntErrors(r.Request.URL.String(), bodyStr)
			errors = append(errors, errorsSlice...)
		}

		// HERE SCAN FOR INFOS
		if scan.InfoFlag && !sliceUtils.Contains(scan.IgnoreExtensions, urlUtils.GetURLExtension(r.Request.URL)) {
			infosSlice := huntInfos(r.Request.URL.String(), bodyStr)
			infos = append(infos, infosSlice...)
		}

		// Colly callbacks run concurrently: add the findings to the shared
		// results all at once, under the lock.
		resultsMu.Lock()

		results.Secrets = append(results.Secrets, secrets...)
		results.Endpoints = append(results.Endpoints, endpoints...)
		results.Extensions = append(results.Extensions, extensionsFound...)
		results.Errors = append(results.Errors, errors...)
		results.Infos = append(results.Infos, infos...)

		resultsMu.Unlock()

		if scan.JSON {
			jsonOutput, err := output.GetJSONString(
				r, secrets, parameters, filetype, errors, infos,
			)
			if err == nil {
				fmt.Println(string(jsonOutput))
			} else {
				log.Println(err)
			}
		}
	})

	// Start scraping on target, probing the sitemap too: it's added to the
	// results only if it's a valid sitemap (see OnResponse).
	path, err := urlUtils.GetPath(target)
	if err == nil && (path == "" || path == "/") {
		sitemapURL := fmt.Sprintf("%s://%s/sitemap.xml", protocolTemp, targetTemp)
		if !ignoreBool || !IgnoreMatch(sitemapURL, &ignoreSlice) {
			probeCtx := colly.NewContext()
			probeCtx.Put(sitemapProbeKey, "true")

			err := c.Request(http.MethodGet, sitemapURL, nil, probeCtx, nil)
			if err != nil && scan.Debug {
				log.Println(err)
			}
		}
	}

	err = c.Visit(target)
	if err != nil && scan.Debug {
		log.Println(err)
	}

	wait()

	// With a max depth, the target is at depth 0, the links found in it at
	// depth 1, the links found in those pages at depth 2 and so on. The
	// links found at a depth are crawled together at the next one (see
	// visitLink): a page is always reached through the fewest links, never
	// first through a longer chain, which would leave its links beyond the
	// max depth.
	for links := event.next.take(); len(links) > 0 && ctx.Err() == nil; links = event.next.take() {
		event.depth++

		for _, link := range links {
			err := c.Visit(link)
			if err != nil && scan.Debug {
				log.Println(err)
			}
		}

		wait()
	}

	// After a CTRL+C the callbacks still running can add to the results:
	// return a copy taken under the lock (appends never change the items
	// already in the copy).
	resultsMu.Lock()
	defer resultsMu.Unlock()

	snapshot := *results

	return &snapshot
}

// CreateColly takes as input all the settings needed to instantiate
// a new Colly Collector object and it returns this object.
func CreateColly(delayTime, concurrency, timeout int,
	cache, rua bool,
	transport http.RoundTripper, userAgent string) *colly.Collector {
	c := colly.NewCollector(
		colly.Async(true),
	)
	c.IgnoreRobotsTxt = true
	c.AllowURLRevisit = false
	// Scan the error pages too (e.g. a 500 with a stack trace): by default
	// colly passes only the successful responses to the callbacks.
	c.ParseHTTPErrorResponse = true

	// Request every URL only once (see visitedStorage).
	err := c.SetStorage(newVisitedStorage())
	if err != nil {
		log.Fatal(err)
	}

	// Only URLs in scope are requested (links and redirects are checked),
	// so the limits can apply to every request.
	err = c.Limit(
		&colly.LimitRule{
			Parallelism: concurrency,
			Delay:       time.Duration(delayTime) * time.Second,
			DomainGlob:  "*",
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

	c.WithTransport(transport)

	return c
}

// newTransport returns the HTTP transport of a crawl, using proxy if set.
// It keeps an idle connection for each parallel request, so connections
// are reused: by default only 2 per host are kept, so most of the
// parallel requests opened a new connection.
func newTransport(concurrency int, proxy string) *http.Transport {
	idleConns := max(concurrency, 1)

	transport := &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        idleConns,
		MaxIdleConnsPerHost: idleConns,
		IdleConnTimeout:     idleConnTimeout,
	}

	// Use a Proxy if needed
	if proxy != "" {
		proxyParsed, err := url.Parse(proxy)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		transport.Proxy = http.ProxyURL(proxyParsed)
	}

	return transport
}

// registerHTMLEvents registers the associated functions for each
// HTML event triggering an action.
func registerHTMLEvents(c *colly.Collector, event *Event) {
	c.OnHTML("a[href], link[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if len(link) != 0 && link[0] != '#' {
			visitHTMLLink(link, event, e)
		}
	})

	c.OnHTML("script[src], iframe[src], svg[src], img[src], video[src], embed[src]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("src"), event, e)
	})

	c.OnHTML("form[action]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("action"), event, e)
	})

	c.OnHTML("[data-src]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("data-src"), event, e)
	})

	c.OnHTML("[data-href]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("data-href"), event, e)
	})

	c.OnHTML("[data-url]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("data-url"), event, e)
	})

	c.OnHTML("img[srcset], source[srcset]", func(e *colly.HTMLElement) {
		srcset := e.Attr("srcset")
		entries := strings.Split(srcset, ",")

		for _, entry := range entries {
			parts := strings.Fields(strings.TrimSpace(entry))
			if len(parts) > 0 {
				link := parts[0]
				visitHTMLLink(link, event, e)
			}
		}
	})

	// Meta refresh redirects, e.g. <meta http-equiv="Refresh" content="0; URL='/next'">.
	c.OnHTML("meta[http-equiv]", func(e *colly.HTMLElement) {
		if refresh, ok := metaRefreshURL(e.Attr("http-equiv"), e.Attr("content")); ok {
			visitHTMLLink(refresh, event, e)
		}
	})

	// The content of <noscript> is parsed as text (as browsers do when
	// JavaScript is enabled): parse it to find the meta refresh redirects
	// for the browsers without JavaScript.
	c.OnHTML("noscript", func(e *colly.HTMLElement) {
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(e.Text))
		if err != nil {
			return
		}

		doc.Find("meta[http-equiv]").Each(func(_ int, meta *goquery.Selection) {
			if refresh, ok := metaRefreshURL(meta.AttrOr("http-equiv", ""), meta.AttrOr("content", "")); ok {
				visitHTMLLink(refresh, event, e)
			}
		})
	})

	c.OnHTML("object[data]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("data"), event, e)
	})

	c.OnHTML("applet[code]", func(e *colly.HTMLElement) {
		visitHTMLLink(e.Attr("code"), event, e)
	})
}

// registerXMLEvents registers the associated functions for each
// XML event triggering an action.
func registerXMLEvents(c *colly.Collector, event *Event) {
	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//url/loc", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e)
	})

	// Create a callback on the XPath query searching for the sitemaps
	// listed in a sitemap index
	c.OnXML("//sitemap/loc", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//link", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//href", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e)
	})

	// Create a callback on the XPath query searching for the URLs
	c.OnXML("//fileurl", func(e *colly.XMLElement) {
		visitXMLLink(e.Text, event, e)
	})
}

// visitHTMLLink checks if the collector should visit a link or not.
// The links of the error pages are not followed: a 404 page with relative
// links, served for any path, would lead to endless /a/a/a/... chains.
func visitHTMLLink(link string, event *Event, e *colly.HTMLElement) {
	if len(link) != 0 && !strings.HasPrefix(link, "data:image") && isSuccess(e.Response) {
		absoluteURL := urlUtils.AbsoluteURL(event.ProtocolTemp, event.TargetTemp, e.Request.AbsoluteURL(link))
		// Visit link found on page
		// Only those links are visited which are in scope
		visitLink(event, e.Request, absoluteURL)
	}
}

// visitXMLLink checks if the collector should visit a link or not
// (not if it's found in an error page, see visitHTMLLink).
func visitXMLLink(link string, event *Event, e *colly.XMLElement) {
	if len(link) != 0 && !strings.HasPrefix(link, "data:image") && isSuccess(e.Response) {
		absoluteURL := urlUtils.AbsoluteURL(event.ProtocolTemp, event.TargetTemp, e.Request.AbsoluteURL(link))
		// Visit link found on page
		// Only those links are visited which are in scope
		visitLink(event, e.Request, absoluteURL)
	}
}

// visitLink is a protocol agnostic wrapper to visit a link
// found in the response to r.
func visitLink(event *Event, r *colly.Request, absoluteURL string) {
	if !canVisit(event, absoluteURL) {
		return
	}

	// With a max depth, the link is crawled at the next depth (see New).
	if event.MaxDepth > 0 {
		if event.depth < event.MaxDepth {
			event.next.add(absoluteURL)
		}

		return
	}

	err := r.Visit(absoluteURL)
	if err != nil && event.Debug {
		log.Println(err)
	}
}

// redirectHandler follows only the redirects in scope and not ignored
// (e.g. http://target.com -> https://target.com), at most maxRedirects.
func redirectHandler(event *Event) func(req *http.Request, via []*http.Request) error {
	return func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects || !canVisit(event, req.URL.String()) {
			// Stop here: the redirect response is handled as an error.
			return http.ErrUseLastResponse
		}

		return nil
	}
}
