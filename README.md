<p align="center">
  <img src="https://github.com/edoardottt/images/blob/main/cariddi/logo.png"><br>
  <b>Take a list of domains, crawl urls and scan for endpoints, secrets, api keys, file extensions, tokens and more</b><br>
  <br>
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

<p align="center">
  <img src="https://github.com/edoardottt/images/blob/main/cariddi/cariddi.gif">
</p>

Installation 📡
----------

#### Homebrew

```console
brew install cariddi
```

#### Snap

```console
sudo snap install cariddi
```

#### Golang

```console
go install -v github.com/edoardottt/cariddi/v2/cmd/cariddi@latest
```

#### Pacman

```console
pacman -Syu cariddi
```

#### NixOS

```console
nix-shell -p cariddi
```

#### Building from source

You need [Go](https://go.dev/) (>=1.26)

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
cariddi -u https://edoardottt.github.io/cariddi-test/
```

With multiple targets you can repeat `-u` or use a file instead, e.g. urls.txt containing:

```console
https://edoardottt.github.io/cariddi-test/
http://testphp.vulnweb.com/
```

```console
cariddi -l urls.txt
cat urls.txt | cariddi
```

On Windows:

```console
.\cariddi.exe -l urls.txt
powershell.exe -Command "cat urls.txt | .\cariddi.exe"
`cat urls.txt | cariddi.exe` # using PowerShell
```

### Basics

- `cariddi -v` (Print the version)
- `cariddi -h` (Print the help)
- `cariddi -h hunt` (Print the help of a group of flags)
- `cariddi -examples` (Print the examples)

### Input

- `cariddi -u https://example.com` (Crawl a URL)
- `cariddi -u https://example.com -u https://example.org` (Crawl more URLs)
- `cariddi -l urls.txt` (Crawl the URLs of a file, one per line)
- `cat urls.txt | cariddi` (Crawl the URLs read from stdin, one per line)

### Scan options

- `cat urls.txt | cariddi -subs` (Crawl searching also subdomains, same as `*.target.com`)
- `cat urls.txt | cariddi -s` (Hunt for secrets)
- `cat urls.txt | cariddi -err` (Hunt for errors in websites)
- `cat urls.txt | cariddi -e` (Hunt for juicy endpoints)
- `cat urls.txt | cariddi -info` (Hunt for useful information in websites)
- `cat urls.txt | cariddi -ext 2` (Hunt for juicy (level 2 out of 7) files)
- `cat urls.txt | cariddi -ef endpoints_file` (Hunt for custom endpoints)
- `cat urls.txt | cariddi -sf secrets_file` (Hunt for custom secrets)
- `cat urls.txt | cariddi -ie pdf,png,jpg` (Ignore these extensions while scanning)

Default: png, svg, jpg, jpeg, bmp, jfif, gif, webp, woff, woff2, ttf, tiff, tif, mp4, webm, mkv, avi, mov, flv, wmv, mp3, wav, flac, ogg, m4a, aac, ico, cur, eot, otf are ignored while scanning for secrets, info and errors.

### Configuration

- `cat urls.txt | cariddi -proxy http://127.0.0.1:8080` (Set a Proxy: http, https, socks5 and socks5h supported)
- `cat urls.txt | cariddi -proxy http://127.0.0.1:8080 -proxy socks5://127.0.0.1:1080` (Use more proxies: each request uses the next one in turn)
- `cat urls.txt | cariddi -proxy proxies.txt` (Use the proxies of a file, one per line, in turn)
- `cat urls.txt | HTTPS_PROXY=http://127.0.0.1:8080 cariddi` (Without -proxy, the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables are used)
- `cat urls.txt | cariddi -d 3` (Follow the links up to 3 levels deep)
- `cat urls.txt | cariddi -rd 2` (2 seconds between a page crawled and another)
- `cat urls.txt | cariddi -t 200` (Send up to 200 concurrent requests)
- `cat urls.txt | cariddi -i forum,blog,community,open` (Ignore urls containing these words)
- `cat urls.txt | cariddi -i ignore_file` (Ignore urls containing at least one line in the input file)
- `cat urls.txt | cariddi -cache` (Use the .cariddi_cache folder as cache)
- `cat urls.txt | cariddi -timeout 5` (Set the timeout for the requests)
- `cat urls.txt | cariddi -H "Cookie: auth=admin;type=2;; X-Custom: customHeader"` (Use custom headers)
- `cat urls.txt | cariddi -H headers.txt` (Read from an external file custom headers)
- `cat urls.txt | cariddi -ua "Custom User Agent"` (Use a custom User Agent)
- `cat urls.txt | cariddi -rua` (Use a random browser user agent on every request)

With more than one proxy, each request uses the next one in turn: the requests are spread among more IPs, but `-t` and `-rd` apply to all of them together and the cookies are shared. Use working proxies: a request through a proxy that fails is lost, with the links of its page (`-debug` shows the requests failed and the proxy used).

### Output

- `cat urls.txt | cariddi -silent` (Print only the URLs found, without the banner)
- `cat urls.txt | cariddi -ot target_name` (Results in txt file)
- `cat urls.txt | cariddi -oh target_name` (Results in html file)
- `cat urls.txt | cariddi -j` (Print the output as JSON lines in stdout)
- `cat urls.txt | cariddi -sr` (Store HTTP responses)
- `cat urls.txt | cariddi -debug` (Print debug information while crawling)

Get Started 🎉
----------

`cariddi -h` prints the help.

```console
Usage:
  cariddi [flags]

Flags:
INPUT:
   -u, -url string[]   URL to crawl (repeat -u to crawl more URLs).
   -l, -list string[]  File containing the URLs to crawl, one per line (repeat -l to read more files).

CONFIGURATION:
   -d, -depth int           Maximum depth to crawl: 1 follows the links of the target, 2 also their links (0 = no limit).
   -timeout int             Set timeout for the requests, in seconds. (default 10)
   -H, -headers string[]    Use custom headers for each request (e.g. -H "Cookie: auth=yes;;Client: type=2"), or read them from a file (one per line). Repeatable.
   -ua, -user-agent string  Use a custom User Agent.
   -rua, -random-agent      Use a random browser user agent on every request.
   -proxy string[]          Set a Proxy to be used (http, https, socks5 and socks5h supported), or read them from a file (one per line). Repeatable: each request uses the next proxy in turn. Default: the HTTP_PROXY, HTTPS_PROXY and NO_PROXY environment variables.
   -cache                   Use the .cariddi_cache folder as cache.

SCOPE:
   -subs, -subdomains    Crawl searching for resources matching the root domain and its subdomains.
   -i, -ignore string[]  Ignore the URLs containing at least one of these comma-separated strings, or of the lines of this file. Repeatable.

RATE-LIMIT:
   -t, -threads int  Number of concurrent requests. (default 20)
   -rd, -delay int   Delay between a page crawled and another, in seconds.

HUNT:
   -s, -secrets                 Hunt for secrets.
   -sf, -secrets-file string    Hunt for secrets with the regexes of this file (one per line) instead of the default ones.
   -e, -endpoints               Hunt for juicy endpoints.
   -ef, -endpoints-file string  Hunt for juicy endpoints with the parameters of this file (one per line) instead of the default ones.
   -ext, -extensions int        Hunt for juicy file extensions. Integer from 1(juicy) to 7(not juicy).
   -err, -errors                Hunt for errors in websites.
   -info                        Hunt for useful information in websites.
   -ie, -ignore-ext string[]    Comma-separated list of extensions to ignore while scanning for secrets, errors and info. (default ["png,svg,jpg,jpeg,bmp,jfif,gif,webp,woff,woff2,ttf,tiff,tif,mp4,webm,mkv,avi,mov,flv,wmv,mp3,wav,flac,ogg,m4a,aac,ico,cur,eot,otf"])

OUTPUT:
   -ot, -output-txt string   Write the output into a TXT file.
   -oh, -output-html string  Write the output into an HTML file.
   -j, -jsonl                Print the output as JSON lines in stdout.
   -sr, -store-response      Store HTTP responses.
   -silent                   Print only the URLs found, without the banner.
   -debug                    Print debug information while crawling.
   -v, -version              Print the version.
   -examples                 Print the examples.
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

Test using [https://edoardottt.github.io/cariddi-test/](https://edoardottt.github.io/cariddi-test/)

```console
cariddi -u https://edoardottt.github.io/cariddi-test/
```

**Help me build this!**

Special thanks to: [go-colly](http://go-colly.org/), [ocervell](https://github.com/ocervell), [zricethezav](https://github.com/gitleaks/gitleaks/blob/master/config/gitleaks.toml), [projectdiscovery](https://github.com/projectdiscovery/nuclei-templates/tree/master/file/keys), [tomnomnom](https://github.com/tomnomnom/gf/tree/master/examples), [RegexPassive](https://github.com/hahwul/RegexPassive) and [all the contributors](https://github.com/edoardottt/cariddi/graphs/contributors).

License 📝
-------

This repository is under [GNU General Public License v3.0](https://github.com/edoardottt/cariddi/blob/main/LICENSE).  
[edoardottt.com](https://edoardottt.com/) to contact me.
