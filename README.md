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
    <img src="https://github.com/edoardottt/cariddi/actions/workflows/go.yml/badge.svg?branch=main" alt="workflows" />
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
  <a href="#installation-">Install</a> •
  <a href="#usage-">Usage</a> •
  <a href="#get-started-">Get Started</a> •
  <a href="#changelog-">Changelog</a> •
  <a href="#contributing-">Contributing</a> •
  <a href="#license-">License</a>
</p>

<!--[![asciicast](https://asciinema.org/a/415989.svg)](https://asciinema.org/a/415989)-->

<p align="center">
  <img src="https://github.com/edoardottt/images/blob/main/cariddi/cariddi.gif">
</p>

Installation 📡
----------

### Homebrew

```console
brew install cariddi
```

### Snap

```console
sudo snap install cariddi
```

### Golang

```console
go install -v github.com/edoardottt/cariddi/cmd/cariddi@latest
```

### Pacman

```console
pacman -Syu cariddi
```

### Building from source

You need [Go](https://go.dev/) (>=1.23)

<details>
  <summary>Building from source for Linux and Windows</summary>

#### Linux

```console
git clone https://github.com/edoardottt/cariddi.git
cd cariddi
go get ./...
make linux # (to install)
make unlinux # (to uninstall)
```

One-liner: `git clone https://github.com/edoardottt/cariddi.git && cd cariddi && go get ./... && make linux`

#### Windows 

Note that the executable works only in cariddi folder.

```console
git clone https://github.com/edoardottt/cariddi.git
cd cariddi
go get ./...
.\make.bat windows # (to install)
.\make.bat unwindows # (to uninstall)
```

</details>

Usage 💡
----------

If you want to scan only a single target you can use

```console
echo https://edoardottt.com/ | cariddi
```

With multiple targets you can use a file instead, e.g. urls.txt containing:

```console
https://edoardottt.com/
http://testphp.vulnweb.com/
```

For Windows:

- use `powershell.exe -Command "cat urls.txt | .\cariddi.exe"` inside the Command prompt
- or just `cat urls.txt | cariddi.exe` using PowerShell

### Basics

- `cariddi -version` (Print the version)
- `cariddi -h` (Print the help)
- `cariddi -examples` (Print the examples)

### Scan options

- `cat urls.txt | cariddi -intensive` (Crawl searching also subdomains, same as `*.target.com`)
- `cat urls.txt | cariddi -s` (Hunt for secrets)
- `cat urls.txt | cariddi -err` (Hunt for errors in websites)
- `cat urls.txt | cariddi -e` (Hunt for juicy endpoints)
- `cat urls.txt | cariddi -info` (Hunt for useful informations in websites)
- `cat urls.txt | cariddi -ext 2` (Hunt for juicy (level 2 out of 7) files)
- `cat urls.txt | cariddi -e -ef endpoints_file` (Hunt for custom endpoints)
- `cat urls.txt | cariddi -s -sf secrets_file` (Hunt for custom secrets)
- `cat urls.txt | cariddi -ie pdf,png,jpg` (Ignore these extensions while scanning)

Default: png, jpg, jpeg, gif, webp, woff, woff2, tiff, tif are ignored by default while scanning for secrets, info and errors.

### Configuration

- `cat urls.txt | cariddi -proxy http://127.0.0.1:8080` (Set a Proxy, http and socks5 supported)
- `cat urls.txt | cariddi -d 2` (2 seconds between a page crawled and another)
- `cat urls.txt | cariddi -c 200` (Set the concurrency level to 200)
- `cat urls.txt | cariddi -i forum,blog,community,open` (Ignore urls containing these words)
- `cat urls.txt | cariddi -it ignore_file` (Ignore urls containing at least one line in the input file)
- `cat urls.txt | cariddi -cache` (Use the .cariddi_cache folder as cache)
- `cat urls.txt | cariddi -t 5` (Set the timeout for the requests)
- `cat urls.txt | cariddi -headers "Cookie: auth=admin;type=2;; X-Custom: customHeader"`
- `cat urls.txt | cariddi -headersfile headers.txt` (Read from an external file custom headers)
- `cat urls.txt | cariddi -ua "Custom User Agent"` (Use a custom User Agent)
- `cat urls.txt | cariddi -rua` (Use a random browser user agent on every request)

### Output

- `cat urls.txt | cariddi -plain` (Print only results)
- `cat urls.txt | cariddi -ot target_name` (Results in txt file)
- `cat urls.txt | cariddi -oh target_name` (Results in html file)
- `cat urls.txt | cariddi -json` (Print the output as JSON in stdout)
- `cat urls.txt | cariddi -sr` (Store HTTP responses)
- `cat urls.txt | cariddi -debug` (Print debug information while crawling)
- `cat urls.txt | cariddi -md 3` (Max 3 depth levels)

Get Started 🎉
----------

`cariddi -h` prints the help.

```console
Usage of cariddi:
  -c int
     Concurrency level. (default 20)
  -cache
     Use the .cariddi_cache folder as cache.
  -d int
     Delay between a page crawled and another.
  -debug
     Print debug information while crawling.
  -e Hunt for juicy endpoints.
  -ef string
     Use an external file (txt, one per line) to use custom parameters for endpoints hunting.
  -err
     Hunt for errors in websites.
  -examples
     Print the examples.
  -ext int
     Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
  -h Print the help.
  -headers string
     Use custom headers for each request E.g. -headers "Cookie: auth=yes;;Client: type=2".
  -headersfile string
     Read from an external file custom headers (same format of headers flag).
  -json
     Print the output as JSON in stdout.
  -md
     Maximum depth level the crawler will follow from the initial target URL.
  -i string
     Ignore the URL containing at least one of the elements of this array.
  -ie value
     Comma-separated list of extensions to ignore while scanning.
  -info
     Hunt for useful informations in websites.
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
  -s Hunt for secrets.
  -sf string
     Use an external file (txt, one per line) to use custom regexes for secrets hunting.
  -sr
     Store HTTP responses.
  -t int
     Set timeout for the requests. (default 10)
  -ua string
     Use a custom User Agent.
  -version
     Print the version.
```

<details>
  <summary>Click to understand <strong>How to integrate cariddi with Burpsuite</strong></summary>

   Normally you use Burpsuite within your browser, so you just have to trust the burpsuite's certificate in the browser and you're done.  
   In order to use cariddi with the BurpSuite proxy you should do some steps further.  

   If you try to use cariddi with the option `-proxy http://127.0.0.1:8080` you will find this error in the burpsuite error log section:  

   ```bash
   Received fatal alert: bad_certificate (or something similar related to the certificate).
   ```

   To make cariddi working fine with Burpsuite you have also to trust the certificate within your entire pc, not just only the browser. These are the steps you have to follow:

   Go to Proxy tab in Bupsuite, then Options. Click on the CA Certificate button and export Certificate in DER format  

   ```bash
   openssl x509 -in burp.der -inform DER -out burp.pem -outform PEM
   sudo chown root:root burp.pem
   sudo chmod 644 burp.pem
   sudo cp burp.pem /usr/local/share/ca-certificates/
   sudo c_rehash
   cd /etc/ssl/certs/
   sudo ln -s /usr/local/share/ca-certificates/burp.pem
   sudo c_rehash .
   ```

   Source: Trust Burp Proxy certificate in Debian/Ubuntu  

   After these steps, in order to use cariddi with Burpsuite you have to:  

   1. Open Burpsuite, making sure that the proxy is listening.  
   2. Use cariddi with the flag `-proxy http://127.0.0.1:8080`.  
   3. You will see that requests and responses will be logged in Burpsuite.

</details>

Changelog 📌
-------

Detailed changes for each release are documented in the [release notes](https://github.com/edoardottt/cariddi/releases).

Contributing 🛠
-------

Just open an [issue](https://github.com/edoardottt/cariddi/issues)/[pull request](https://github.com/edoardottt/cariddi/pulls).

Before opening a pull request, download [golangci-lint](https://golangci-lint.run/usage/install/) and run

```console
golangci-lint run
```

If there aren't errors, go ahead :)

**Help me build this!**

Special thanks to: [go-colly](http://go-colly.org/), [ocervell](https://github.com/ocervell), [zricethezav](https://github.com/gitleaks/gitleaks/blob/master/config/gitleaks.toml), [projectdiscovery](https://github.com/projectdiscovery/nuclei-templates/tree/master/file/keys), [tomnomnom](https://github.com/tomnomnom/gf/tree/master/examples), [RegexPassive](https://github.com/hahwul/RegexPassive) and [all the contributors](https://github.com/edoardottt/cariddi/graphs/contributors).

License 📝
-------

This repository is under [GNU General Public License v3.0](https://github.com/edoardottt/cariddi/blob/main/LICENSE).  
[edoardottt.com](https://edoardottt.com/) to contact me.
