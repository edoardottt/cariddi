<p align="center">
  <img src="https://github.com/edoardottt/images/blob/main/cariddi/logo.png"><br>
  <b>Take a list of domains, crawl urls and scan for endpoints, secrets, api keys, file extensions, tokens and more</b><br>
  <br>
  <!-- go-report-card -->
  <a href="https://goreportcard.com/report/github.com/edoardottt/cariddi">
    <img src="https://goreportcard.com/badge/github.com/edoardottt/cariddi" alt="go-report-card" />
  </a>
  <!-- workflows -->
  <a href="https://github.com/edoardottt/cariddi/actions">
    <img src="https://github.com/edoardottt/cariddi/workflows/Go/badge.svg?branch=main" alt="workflows" />
  </a>
  <!-- ubuntu-build -->
  <a href="https://edoardoottavianelli.it">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/ubuntu-build.svg" alt="ubuntu-build" />
  </a>
  <!-- win10-build -->
  <a href="https://edoardoottavianelli.it">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/win10.svg" alt="win10-build" />
  </a>
  <!-- pr-welcome -->
  <a href="https://edoardoottavianelli.it">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/pr-welcome.svg" alt="pr-welcome" />
  </a>

  <br>
  
  <!-- mainteinance -->
  <a href="https://edoardoottavianelli.it">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/maintained-yes.svg" alt="Mainteinance yes" />
  </a>
  <!-- ask-me-anything -->
  <a href="https://edoardoottavianelli.it">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/ask-me-anything.svg" alt="ask me anything" />
  </a>
  <!-- gobadge -->
  <a href="https://edoardoottavianelli.it">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/gobadge" alt="gobadge" />
  </a>
  <!-- license GPLv3.0 -->
  <a href="https://github.com/edoardottt/cariddi/blob/master/LICENSE">
    <img src="https://github.com/edoardottt/images/blob/main/cariddi/license-GPL3.svg" alt="license-GPL3" />
  </a>
  <br>
  <sub>
    Coded with 💙 by edoardottt
  </sub>
  <br>
  <!--Tweet button-->
  <a href="https://twitter.com/intent/tweet?url=https://github.com/edoardottt/cariddi&text=Take%20a%20list%20of%20domains,%20crawl%20urls%20and%20scan%20for%20endpoints,%20secrets,%20api%20keys,%20file%20extensions,%20tokens%20and%20more...%20%23network%20%23security%20%23infosec%20%23oss%20%23github%20%23bugbounty%20%23linux" target="_blank">Share on Twitter!
  </a>
</p>
<p align="center">
  <a href="#preview-bar_chart">Preview</a> •
  <a href="#installation-">Install</a> •
  <a href="#get-started-">Get Started</a> •
  <a href="#examples-">Examples</a> •
  <a href="#changelog-">Changelog</a> •
  <a href="#contributing-">Contributing</a> •
  <a href="#license-">License</a>
</p>

Preview :bar_chart:
----------

<!--[![asciicast](https://asciinema.org/a/415989.svg)](https://asciinema.org/a/415989)-->

<p align="center">
  <img src="https://github.com/edoardottt/images/blob/main/cariddi/cariddi.gif">
</p>

Installation 📡
----------

### Using Snap
```bash
sudo snap install cariddi
```

### Using Go
```bash
go install -v github.com/edoardottt/cariddi/cmd/cariddi@latest
```

### Building from source

You need [Go](https://golang.org/).

- **Linux**

  - `git clone https://github.com/edoardottt/cariddi.git`
  - `cd cariddi`
  - `go get`
  - `make linux` (to install)
  - `make unlinux` (to uninstall)

  Or in one line: `git clone https://github.com/edoardottt/cariddi.git; cd cariddi; go get; make linux`

- **Windows** (executable works only in cariddi folder.)

  - `git clone https://github.com/edoardottt/cariddi.git`
  - `cd cariddi`
  - `go get`
  - `.\make.bat windows` (to install)    
  - `.\make.bat unwindows` (to uninstall)

Get Started 🎉
----------

`cariddi -h` prints the help in the command line.

*Note*: Don't rely on the CLI output, use always `-ot/-oh` to save the output.

```
Usage of cariddi:
  -c int
    	Concurrency level. (default 20)
  -cache
    	Use the .cariddi_cache folder as cache.
  -d int
    	Delay between a page crawled and another.
  -debug
    	Print debug information while crawling.
  -e	Hunt for juicy endpoints.
  -ef string
    	Use an external file (txt, one per line) to use custom parameters for endpoints hunting.
  -err
    	Hunt for errors in websites.
  -examples
    	Print the examples.
  -ext int
    	Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
  -h	Print the help.
  -headers string
    	Use custom headers for each request E.g. -headers "Cookie: auth=yes;;Client: type=2".
  -headersfile string
    	Read from an external file custom headers (same format of headers flag).
  -i string
    	Ignore the URL containing at least one of the elements of this array.
  -info
    	Hunt for useful informations in websites.
  -insecure
    	Ignore invalid HTTPS certificates.
  -intensive
    	Crawl searching for resources matching 2nd level domain.
  -it string
    	Ignore the URL containing at least one of the lines of this file.
  -oh string
    	Write the output into an HTML file.
  -ot string
    	Write the output into a TXT file.
  -plain
    	Print only the results.
  -proxy string
    	Set a Proxy to be used (http and socks5 supported).
  -rua
    	Use a random browser user agent on every request.
  -s	Hunt for secrets.
  -sf string
    	Use an external file (txt, one per line) to use custom regexes for secrets hunting.
  -t int
    	Set timeout for the requests. (default 10)
  -ua string
    	Use a custom User Agent.
  -version
    	Print the version.
```


Examples 💡
----------

  - `cariddi -version` (Print the version)
  - `cariddi -h` (Print the help)
  - `cariddi -examples` (Print the examples)
  - `cat urls | cariddi -s` (Hunt for secrets)
  - `cat urls | cariddi -d 2` (2 seconds between a page crawled and another)
  - `cat urls | cariddi -c 200` (Set the concurrency level to 200)
  - `cat urls | cariddi -e` (Hunt for juicy endpoints)
  - `cat urls | cariddi -plain` (Print only useful things)
  - `cat urls | cariddi -ot target_name` (Results in txt file)
  - `cat urls | cariddi -oh target_name` (Results in html file)
  - `cat urls | cariddi -ext 2` (Hunt for juicy (level 2 out of 7) files)
  - `cat urls | cariddi -e -ef endpoints_file` (Hunt for custom endpoints)
  - `cat urls | cariddi -s -sf secrets_file` (Hunt for custom secrets)
  - `cat urls | cariddi -i forum,blog,community,open` (Ignore urls containing these words)
  - `cat urls | cariddi -it ignore_file` (Ignore urls containing at least one line in the input file)
  - `cat urls | cariddi -cache` (Use the .cariddi_cache folder as cache)
  - `cat urls | cariddi -t 5` (Set the timeout for the requests)
  - `cat urls | cariddi -intensive` (Crawl searching also subdomains, same as `*.target.com`)
  - `cat urls | cariddi -rua` (Use a random browser user agent on every request)
  - `cat urls | cariddi -proxy http://127.0.0.1:8080` (Set a Proxy (http and socks5 supported))
  - `cat urls | cariddi -headers "Cookie: auth=admin;type=2;; X-Custom: customHeader"`
  - `cat urls | cariddi -headersfile headers.txt` (Read from an external file custom headers)
  - `cat urls | cariddi -err` (Hunt for errors in websites)
  - `cat urls | cariddi -info` (Hunt for useful informations in websites)
  - `cat urls | cariddi -debug` (Print debug information while crawling)
  - `cat urls | cariddi -ua "Custom User Agent"` (Use a custom User Agent)
  - `cat urls | cariddi -insecure` (Ignore invalid HTTPS certificates)

  - For Windows:
  	- use `powershell.exe -Command "cat urls | .\cariddi.exe"` inside the Command prompt
  	- or just `cat urls | cariddi.exe` using PowerShell

  - To integrate cariddi with Burpsuite [make sure to follow these steps](https://github.com/edoardottt/cariddi/wiki/BurpSuite-Integration).

Changelog 📌
-------
Detailed changes for each release are documented in the [release notes](https://github.com/edoardottt/cariddi/releases).

Contributing 🛠
-------

Just open an [issue](https://github.com/edoardottt/cariddi/issues)/[pull request](https://github.com/edoardottt/cariddi/pulls).

Before opening a pull request, download [golangci-lint](https://golangci-lint.run/usage/install/) and run
```bash
golangci-lint run
```
If there aren't errors, go ahead :)

**Help me building this!**

Special thanks to: [go-colly](http://go-colly.org/), [zricethezav](https://github.com/zricethezav/gitleaks/blob/master/config/default.go), [projectdiscovery](https://github.com/projectdiscovery/nuclei-templates/tree/master/file/keys), [tomnomnom](https://github.com/tomnomnom/gf/tree/master/examples) and [RegexPassive](https://github.com/hahwul/RegexPassive).

**To do:**

  - [ ] Add more Tests
  
  - [ ] Tor support
  
License 📝
-------

This repository is under [GNU General Public License v3.0](https://github.com/edoardottt/cariddi/blob/main/LICENSE).  
[edoardoottavianelli.it](https://www.edoardoottavianelli.it) to contact me.
