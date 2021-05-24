<p align="center">
  <img src="https://github.com/edoardottt/images/blob/main/cariddi/logo.png"><br>
  <b>Take a list of domains and scan for endpoints, secrets, api keys, file extensions, tokens and more...</b><br>
  <br>
  <!-- go-report-card -->
  <a href="https://goreportcard.com/report/github.com/edoardottt/cariddi">
    <img src="https://goreportcard.com/badge/github.com/edoardottt/cariddi" alt="go-report-card" />
  </a>
  <!-- workflows -->
  <a href="https://edoardoottavianelli.it">
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
    Coded with 💙 by edoardottt.
  </sub>
  <br>
  <!--Tweet button-->
  <a href="https://twitter.com/intent/tweet?text=Take%20a%20list%20of%20domains%20and%20scan%20for%20endpoints%2C%20secrets%2C%20api%20keys%2C%20file%20extensions%2C%20tokens%20and%20more...%20%23linux%20%23infosec%20%23bugbounty%20%23security%20%23golang%20%23github%20%23oss%20https%3A//github.com/edoardottt/cariddi" target="_blank">Share on Twitter!
  </a>
</p>
<p align="center">
  <a href="#preview-bar_chart">Preview</a> •
  <a href="#installation-">Install</a> •
  <a href="#get-started-">Get Started</a> •
  <a href="#examples-">Examples</a> •
  <a href="#contributing-">Contributing</a>
</p>

Preview :bar_chart:
----------

[![asciicast](https://asciinema.org/a/415989.svg)](https://asciinema.org/a/415989)

Installation 📡
----------

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

`cariddi help` prints the help in the command line.

```
Usage of cariddi:
  -c int
    	Concurrency level. (default 20)
  -d int
    	Delay between a page crawled and another.
  -e	Hunt for juicy endpoints.
  -ef string
    	Use an external file (txt, one per line) to use custom parameters for endpoints hunting.
  -examples
    	Print the examples.
  -ext int
    	Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
  -h	Print the help.
  -oh string
    	Write the output into an HTML file.
  -ot string
    	Write the output into a TXT file.
  -plain
    	Print only the results.
  -s	Hunt for secrets.
  -sf string
    	Use an external file (txt, one per line) to use custom regexes for secrets hunting.
  -version
    	Print the version.
```


Examples 💡
----------

  - `cat urls | cariddi -version` (Print the version)
  - `cat urls | cariddi -h` (Print the help)
  - `cat urls | cariddi -e` (Hunt for secrets)
  - `cat urls | cariddi -d 2` (2 seconds between a page crawled and another)
  - `cat urls | cariddi -c 200` (Set the concurrency level to 200)
  - `cat urls | cariddi -s` (Hunt for juicy endpoints)
  - `cat urls | cariddi -plain` (Print only useful things)
  - `cat urls | cariddi -ot target_name` (Results in txt file)
  - `cat urls | cariddi -oh target_name` (Results in html file)
  - `cat urls | cariddi -ext 2` (Hunt for juicy (level 2 of 7) files)
  - `cat urls | cariddi -e -ef endpoints_file` (Hunt for custom endpoints)
  - `cat urls | cariddi -s -sf secrets_file` (Hunt for custom secrets)

  - For Windows see [here](https://stackoverflow.com/questions/14574170/how-do-i-use-a-pipe-to-redirect-the-output-of-one-command-to-the-input-of-anothe) 

Contributing 🛠
-------

Just open an issue/pull request. See also [CONTRIBUTING.md](https://github.com/edoardottt/cariddi/blob/master/CONTRIBUTING.md) and [CODE OF CONDUCT.md](https://github.com/edoardottt/cariddi/blob/master/CODE_OF_CONDUCT.md)

**Help me building this!**

A special thanks to:

  - [zricethezav](https://github.com/zricethezav/gitleaks/blob/master/config/default.go)

**To do:**

  - [ ] Tests (😂)
  
  - [ ] Tor support
  
  - [ ] Proxy support

  - [x] Plain output (print only results)
  
  - [x] HTML output
  
  - [x] Build an Input Struct and use it as parameter

  - [x] Output color

  - [x] Endpoints (parameters) scan

  - [x] Secrets scan

  - [x] Extensions scan
  
  - [x] TXT output
  
License 📝
-------

This repository is under [GNU General Public License v3.0](https://github.com/edoardottt/cariddi/blob/main/LICENSE).  
[edoardoottavianelli.it](https://www.edoardoottavianelli.it) to contact me.
