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

package scanner

import (
	"regexp"
	"slices"
	"strings"
	"sync"

	"golang.org/x/net/publicsuffix"
)

// Info struct.
// Name = the name that identifies the information.
// Regex = The regular expression to be matched.
// FalsePositives = A list of known false positives (lowercase): a match
// containing one of them, ignoring case, is not reported.
// FalsePositive = if not nil, it reports whether the match
// text[start:end] is a false positive, given its context.
// Sources = links to the official documentation of the format, if any.
type Info struct {
	Name           string
	Regex          regexp.Regexp
	FalsePositives []string
	FalsePositive  func(text string, start, end int) bool
	Sources        []string
}

// InfoMatched struct.
// Info = Info struct.
// Url = url in which the information is found.
// Match = the string matching the regex.
type InfoMatched struct {
	Info  Info
	URL   string
	Match string
}

// InfoScanner hunts for infos in response bodies (see Engine).
// It is safe for concurrent use.
type InfoScanner struct {
	infos  []Info
	engine *Engine
}

const (
	// sshHost is the host of an SSH location: a name or an IPv4 address.
	sshHost = `[A-Za-z0-9](?:[A-Za-z0-9.-]*[A-Za-z0-9])?`
	// sshPath is the path of an SSH location in the scp-like syntax, after the
	// colon: a path with a slash (org/repo, /srv/repo.git, ~/repo) or a .git name.
	sshPath = `(?:[/~]?[\w.~-]*(?:/[\w.~-]*[\w~-])+|[/~]?[\w.-]*\.git\b)`
	// userinfoChars are the bytes of the userinfo of a URL (RFC 3986).
	userinfoChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._~%!$&'()*+,;=:"
	// maxUserinfo is the longest userinfo looked for before an email address.
	maxUserinfo = 256
	// cloudinaryCloudName is a Cloudinary cloud name, quoted or alone up to the end
	// of the line: not a variable, as in cloud_name: cloudName,.
	cloudinaryCloudName = `[A-Za-z0-9][\w-]{1,63}(?:["'\x60]|[ \t]{0,5}(?:[\r\n]|$))`
	// snowflakeAccount is an account identifier of Snowflake: the organization name,
	// - and the account name (also of a private connectivity), or an account
	// locator, also with its region and cloud.
	snowflakeAccount = `(?:[A-Za-z][A-Za-z0-9]{0,40}-\w{1,60}(?:\.privatelink)?|[A-Za-z]{2}\d{5}(?:\.[A-Za-z0-9-]{2,30}){0,3})(?:[^\w.-]|$)`
)

var (
	infos           []Info       //nolint: gochecknoglobals
	onceInfos       sync.Once    //nolint: gochecknoglobals
	infoScanner     *InfoScanner //nolint: gochecknoglobals
	onceInfoScanner sync.Once    //nolint: gochecknoglobals
	// escapeSuffix matches an escape sequence of a JSON or JavaScript
	// string at the end of a text, e.g. \n, \t, \u003e or \x3e.
	escapeSuffix = regexp.MustCompile(`\\(?:[nrt]|u[0-9a-fA-F]{4}|x[0-9a-fA-F]{2})$`)
	// sshPathPrefix matches the :path of an SSH location at the start of a text.
	sshPathPrefix = regexp.MustCompile(`^:` + sshPath)
)

// GetInfoRegexes returns all the info structs.
func GetInfoRegexes() []Info {
	onceInfos.Do(func() {
		infos = []Info{
			{
				// Dot-separated name parts, then domain labels (no - at their ends) and
				// an alphabetic or punycode top-level domain.
				Name: "Email address",
				Regex: *regexp.MustCompile(`[a-zA-Z0-9_%+-]+(?:\.[a-zA-Z0-9_%+-]+)*@(?:[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)+` +
					`(?:xn--[a-zA-Z0-9-]{1,59}|[a-zA-Z]{2,63})\b`),
				FalsePositive: emailFalsePositive,
				Sources: []string{
					"https://www.rfc-editor.org/rfc/rfc5322#section-3.4.1",
					"https://www.rfc-editor.org/rfc/rfc1035#section-2.3.1",
				},
			},
			{
				// As browsers parse them: any content, closed by the first
				// "-->" or "--!>", and the empty "<!-->" and "<!--->".
				Name:    "HTML comment",
				Regex:   *regexp.MustCompile(`(?s)<!--(?:-?>|.*?--!?>)`),
				Sources: []string{"https://html.spec.whatwg.org/multipage/syntax.html#comments"},
			},
			{
				Name:          "IPv4 address",
				Regex:         *regexp.MustCompile(`(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(?:\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}\b`),
				FalsePositive: ipv4FalsePositive,
			},
			{
				// SSH locations of Git repositories: ssh:// URLs and the scp-like syntax
				// (user@host:path, e.g. git@github.com:org/repo.git), with a domain or IP
				// address, or a .git path.
				Name: "Git SSH URL",
				Regex: *regexp.MustCompile(`\b(?:git\+)?ssh(?:\+git)?://(?:[\w.%~:-]+@)?` + sshHost + `(?::\d{1,5})?(?:/[\w.~-]*[\w~-])+|` +
					`\b[A-Za-z0-9_][\w.-]*@(?:` + sshHost + `\.` + sshHost + `:` + sshPath + `|[A-Za-z0-9-]+:[/~]?[\w.~/-]*\.git\b)`),
				FalsePositive: gitSSHFalsePositive,
				Sources:       []string{"https://git-scm.com/docs/git-clone#_git_urls"},
			},
			{
				Name:    "Adobe Client ID",
				Regex:   *regexp.MustCompile(`(?i:adobe[\w.-]{0,20}?(?:client_?id|api_?key))[\w.-]{0,10}` + assign + `[0-9a-f]{32}\b`),
				Sources: []string{"https://experienceleague.adobe.com/en/docs/experience-manager-learn/cloud-service/developing/extensibility/app-builder/server-to-server-auth"},
			},
			{
				// 10 characters assigned to an Algolia or DocSearch application ID name
				// (also the X-Algolia-Application-Id header), or in the hosts of the API.
				Name: "Algolia Application ID",
				Regex: *regexp.MustCompile(keyed(`algolia|docsearch`, `app(?:lication)?[_-]?id`, `[A-Z0-9]{10}\b`) +
					`|\b[A-Za-z0-9]{10}(?:-dsn|-\d)?\.algolia(?:\.net|net\.com)\b`),
				Sources: []string{"https://www.algolia.com/doc/rest-api/search/"},
			},
			{
				// Virtual-hosted (bucket.oss-region.aliyuncs.com) and path-style URLs,
				// and oss:// URIs.
				Name: "Alibaba OSS Bucket",
				Regex: *regexp.MustCompile(`\b[a-z0-9][a-z0-9-]{1,61}[a-z0-9]\.oss-[a-z0-9-]{2,40}\.aliyuncs\.com\b|` +
					`(?:https?:)?\\?/\\?/oss-[a-z0-9-]{2,40}\.aliyuncs\.com\\?/[a-z0-9][a-z0-9-]{1,61}[a-z0-9]|\boss://[a-z0-9][a-z0-9-]{1,61}[a-z0-9]`),
				Sources: []string{"https://www.alibabacloud.com/help/en/oss/user-guide/regions-and-endpoints"},
			},
			{
				// Assigned to an Anthropic or Claude organization ID name (e.g.
				// ANTHROPIC_ORGANIZATION_ID or the anthropic-organization-id header),
				// or to forceLoginOrgUUID of Claude Code (also a list).
				Name: "Anthropic Organization ID",
				Regex: *regexp.MustCompile(`(?:(?i:(?:anthropic|claude)[\w.-]{0,20}?org(?:anization)?[_-]?(?:id|uuid))[\w.-]{0,10}` + assign +
					`|(?i:\bforceLoginOrgUUID)` + assign + `(?:\[\s{0,5}["']?)?)` +
					`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`),
				Sources: []string{
					"https://platform.claude.com/docs/en/api/overview",
					"https://code.claude.com/docs/en/authentication",
				},
			},
			{
				// Tagged IDs: 01 and 22 base58 characters.
				Name:    "Anthropic Workspace ID",
				Regex:   *regexp.MustCompile(`\bwrkspc_01[1-9A-HJ-NP-Za-km-z]{22}\b`),
				Sources: []string{"https://platform.claude.com/docs/en/api/overview"},
			},
			{
				// Workload identity federation issuer, rule and service account IDs:
				// with the organization ID, the configuration of a workload.
				Name:    "Anthropic Federation ID",
				Regex:   *regexp.MustCompile(`\b(?:fdis|fdrl|svac)_01[1-9A-HJ-NP-Za-km-z]{22}\b`),
				Sources: []string{"https://platform.claude.com/docs/en/manage-claude/wif-admin-api"},
			},
			{
				Name:  "Asana Client ID",
				Regex: *regexp.MustCompile(`(?i:asana[\w.-]{0,20}?client_?id)[\w.-]{0,10}` + assign + `\d{16}\b`),
			},
			{
				// Jira, Confluence and the other Atlassian Cloud products of a site.
				Name:    "Atlassian Cloud Site",
				Regex:   *regexp.MustCompile(`\b[A-Za-z0-9][A-Za-z0-9-]{0,62}\.atlassian\.net\b`),
				Sources: []string{"https://developer.atlassian.com/cloud/jira/platform/basic-auth-for-rest-apis/"},
			},
			{
				// 32 characters assigned to an Auth0 client ID name: public, its secret
				// is a secret.
				Name:    "Auth0 Client ID",
				Regex:   *regexp.MustCompile(`(?i:auth0[\w.-]{0,20}?client_?id)[\w.-]{0,10}` + assign + `[A-Za-z0-9]{32}\b`),
				Sources: []string{"https://auth0.com/docs/get-started/applications/application-settings"},
			},
			{
				// Virtual-hosted (bucket.s3[.region].amazonaws.com) and path-style
				// (//s3[.region].amazonaws.com/bucket, also with JSON-escaped slashes)
				// URLs, and s3:// URIs.
				Name: "AWS S3 Bucket",
				Regex: *regexp.MustCompile(`\b[a-z0-9][a-z0-9.-]+\.s3[a-z0-9.-]*\.amazonaws\.com(?:\.cn)?|` +
					`(?:https?:)?\\?/\\?/s3[a-z0-9.-]*\.amazonaws\.com(?:\.cn)?\\?/[a-zA-Z0-9._-]{3,}|` +
					`\bs3[an]?://[a-zA-Z0-9._-]{3,}`),
				Sources: []string{
					"https://docs.aws.amazon.com/AmazonS3/latest/userguide/VirtualHosting.html",
					"https://docs.aws.amazon.com/AmazonS3/latest/userguide/bucketnamingrules.html",
				},
			},
			{
				// ARNs of resources owned by an account, and of S3 buckets.
				Name: "AWS ARN",
				Regex: *regexp.MustCompile(`\barn:aws(?:-[a-z]+){0,2}:[a-z0-9-]{2,40}:[a-z0-9-]*:\d{12}:[\w/:.+=@,*-]*[\w/+=@*-]|` +
					`\barn:aws(?:-[a-z]+){0,2}:s3:::[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]`),
				Sources: []string{"https://docs.aws.amazon.com/IAM/latest/UserGuide/reference-arns.html"},
			},
			{
				// Identity pool IDs: region:GUID.
				Name: "AWS Cognito pool",
				Regex: *regexp.MustCompile(`\b[a-z]{2,4}(?:-[a-z]{2,4})?-(?:east|west|north|south|central|northeast|northwest|southeast|` +
					`southwest)-\d:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`),
				Sources: []string{"https://docs.aws.amazon.com/cognitoidentity/latest/APIReference/API_DescribeIdentityPool.html"},
			},
			{
				// Public by design: it identifies the clients of a GraphQL API.
				Name:    "AWS AppSync API Key",
				Regex:   *regexp.MustCompile(`\bda2-[a-z0-9]{26}\b`),
				Sources: []string{"https://docs.aws.amazon.com/appsync/latest/APIReference/API_ApiKey.html"},
			},
			{
				Name:    "AWS AppSync Endpoint",
				Regex:   *regexp.MustCompile(`\b[a-z0-9]{26}\.appsync-api\.[a-z0-9-]{9,25}\.amazonaws\.com\b`),
				Sources: []string{"https://docs.aws.amazon.com/appsync/latest/devguide/security-authz.html"},
			},
			{
				Name:    "AWS Account ID",
				Regex:   *regexp.MustCompile(`(?i:aws[_.-]?account[_.-]?id)` + assign + `\d{12}\b`),
				Sources: []string{"https://docs.aws.amazon.com/accounts/latest/reference/manage-acct-identifiers.html"},
			},
			{
				// Users, groups, roles, managed policies and their versions, instance
				// profiles, public keys and certificates: the access key IDs are secrets.
				Name:    "AWS IAM Unique ID",
				Regex:   *regexp.MustCompile(`\b(?:AIDA|AGPA|AROA|ANPA|ANVA|AIPA|APKA|ASCA)[A-Z0-9]{16,17}\b`),
				Sources: []string{"https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html"},
			},
			{
				Name:    "Azure Blob/ADLS Storage",
				Regex:   *regexp.MustCompile(`(?i)(?:https?:\/\/)?([a-z0-9]{3,24})\.(blob|dfs)\.core\.windows\.net`),
				Sources: []string{"https://learn.microsoft.com/en-us/azure/storage/common/storage-account-overview"},
			},
			{
				// AI services, Cosmos DB, AI Search, API Management, Container Registry
				// and Batch accounts, and the APIs of function apps.
				Name: "Azure Service Endpoint",
				Regex: *regexp.MustCompile(`\b[a-z0-9][a-z0-9-]{1,62}\.(?:openai\.azure\.com|cognitiveservices\.azure\.com|` +
					`documents\.azure\.com|table\.cosmos\.azure\.com|search\.windows\.net|(?:scm\.|management\.)?azure-api\.net|` +
					`azurecr\.io|[a-z0-9-]{1,30}\.batch\.azure\.com)\b|\b[a-z0-9][a-z0-9-]{1,59}\.azurewebsites\.net/api/`),
				Sources: []string{
					"https://learn.microsoft.com/en-us/azure/ai-foundry/openai/reference",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-cognitive-service-key",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-cosmos-db-account-access-key",
					"https://learn.microsoft.com/en-us/azure/search/search-create-service-portal",
					"https://learn.microsoft.com/en-us/rest/api/apimanagement/apimanagementrest/azure-api-management-rest-api-authentication",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-container-registry-access-key",
					"https://learn.microsoft.com/en-us/azure/azure-functions/functions-bindings-http-webhook-trigger",
				},
			},
			{
				// Initial domains, IDs in the sign-in URLs, and IDs assigned to an
				// Azure tenant ID name.
				Name: "Azure Tenant",
				Regex: *regexp.MustCompile(`(?i)\b[a-z0-9][a-z0-9.-]{0,62}\.onmicrosoft\.com\b|` +
					`\b(?:login\.microsoftonline\.com|login\.windows\.net|sts\.windows\.net)/` +
					`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b|azure[\w.-]{0,10}?tenant_?id` + assign +
					`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`),
				Sources: []string{"https://learn.microsoft.com/en-us/entra/identity-platform/v2-protocols-oidc"},
			},
			{
				Name:    "Azure DevOps Organization",
				Regex:   *regexp.MustCompile(`\bdev\.azure\.com/[A-Za-z0-9][A-Za-z0-9-]{0,49}`),
				Sources: []string{"https://learn.microsoft.com/en-us/azure/devops/extend/develop/work-with-urls"},
			},
			{
				// Shared access signatures granting only read, list and filter.
				Name: "Azure SAS Token (read-only)",
				Regex: *regexp.MustCompile(`\bsp=[flr]{1,3}(?:&|\\u0026)[^\s"'<>]{0,500}?(?:&(?:amp;)?|\\u0026)` +
					`sig=[A-Za-z0-9%/+]{40,100}`),
				Sources: []string{"https://learn.microsoft.com/en-us/rest/api/storageservices/create-service-sas"},
			},
			{
				Name: "BitBucket Client ID",
				Regex: *regexp.MustCompile(`(?i:bitbucket[\w.-]{0,20}?(?:client_?id|consumer_?key|oauth_?key))[\w.-]{0,10}` + assign +
					`[A-Za-z0-9]{18,32}\b`),
				Sources: []string{"https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/"},
			},
			{
				// 16 characters assigned to a merchant ID name within 300 characters after
				// "braintree" (e.g. in the gateway configuration).
				Name:    "Braintree Merchant ID",
				Regex:   *regexp.MustCompile(`(?i:braintree[\s\S]{0,300}?merchant_?id)[\w.-]{0,10}` + assign + `[0-9a-z]{16}\b`),
				Sources: []string{"https://developer.paypal.com/braintree/docs/start/hello-server"},
			},
			{
				// In R2 endpoints (also of a jurisdiction), API and dashboard URLs, or
				// assigned to a Cloudflare account ID name or to CF_ACCOUNT_ID
				// (deprecated Wrangler name): account tokens need it.
				Name: "Cloudflare Account ID",
				Regex: *regexp.MustCompile(`\b[0-9a-f]{32}\.(?:[a-z]{2,10}\.)?r2\.cloudflarestorage\.com\b|` +
					`\b(?:api\.cloudflare\.com/client/v4/accounts|dash\.cloudflare\.com)/[0-9a-f]{32}\b|` +
					`(?:(?i:cloudflare[\w.-]{0,20}?account_?id)[\w.-]{0,10}|(?i:\bcf_account_id))` + assign + `[0-9a-f]{32}\b`),
				Sources: []string{
					"https://developers.cloudflare.com/fundamentals/account/find-account-and-zone-ids/",
					"https://developers.cloudflare.com/r2/reference/data-location/",
					"https://developers.cloudflare.com/workers/wrangler/system-environment-variables/",
				},
			},
			{
				// Assigned to a Cloudinary cloud name name (CLOUDINARY_CLOUD_NAME), also in an
				// object or a configuration call, quoted or alone up to the end of the line
				// (not a variable): public, it is in the URLs of the assets.
				Name: "Cloudinary Cloud Name",
				Regex: *regexp.MustCompile(keyed(`cloudinary`, `cloud[_-]?name`, cloudinaryCloudName) + `|` +
					configCall(`cloudinary`, `cloud_?name`, cloudinaryCloudName)),
				Sources: []string{
					"https://cloudinary.com/documentation/node_integration",
					"https://cloudinary.com/documentation/solution_overview",
				},
			},
			{
				// organizations/, the organization ID, /apiKeys/ and the key ID (also with
				// JSON-escaped slashes): the name of a CDP API key, its private key is a secret.
				Name:    "Coinbase CDP API Key Name",
				Regex:   *regexp.MustCompile(`\borganizations\\?/` + uuid + `\\?/apiKeys\\?/` + uuid + `\b`),
				Sources: []string{"https://docs.cdp.coinbase.com/coinbase-app/authentication-authorization/api-key-authentication"},
			},
			{
				// The ID of an API key: its secret is a secret.
				Name:  "Confluent API Key",
				Regex: *regexp.MustCompile(`(?i:confluent[\w.-]{0,20}?(?:api_?)?key)[\w.-]{0,10}` + assign + `[A-Z0-9]{16}\b`),
				Sources: []string{
					"https://docs.confluent.io/cloud/current/security/authenticate/workload-identities/service-accounts/api-keys/overview.html",
				},
			},
			{
				// pub and 32 hex digits (the format is not documented): public by design,
				// it sends data from browsers and mobile apps.
				Name:    "Datadog Client Token",
				Regex:   *regexp.MustCompile(`\bpub[0-9a-f]{32}\b`),
				Sources: []string{"https://docs.datadoghq.com/account_management/api-app-keys/#client-tokens"},
			},
			{
				// A server in a connection string made of properties (key=value;) that
				// uses the account of the process instead of a password (Integrated Security
				// or Trusted_Connection): it names a server, and often a database. Local and
				// documentation servers are skipped (see propFalsePositive).
				Name: "Database Connection String (Integrated Security)",
				Regex: *regexp.MustCompile(`\b` + propServer + `[ \t]{0,3};` + propPairs + propTrusted +
					`|\b` + propTrusted + `[ \t]{0,3};` + propPairs + propServer),
				FalsePositives: []string{"password=", "pwd="},
				FalsePositive:  func(text string, start, end int) bool { return propFalsePositive(text[start:end]) },
				Sources:        []string{"https://learn.microsoft.com/en-us/dotnet/framework/data/adonet/connection-string-syntax"},
			},
			{
				// Virtual-hosted (bucket.region[.cdn].digitaloceanspaces.com) and
				// path-style URLs.
				Name: "DigitalOcean Spaces",
				Regex: *regexp.MustCompile(`\b[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]\.[a-z]{3}\d\.(?:cdn\.)?digitaloceanspaces\.com\b|` +
					`(?:https?:)?\\?/\\?/[a-z]{3}\d\.digitaloceanspaces\.com\\?/[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]`),
				Sources: []string{"https://docs.digitalocean.com/products/spaces/reference/aws-sdks/"},
			},
			{
				// Assigned to a Discord client or application ID name, or in the authorization
				// URL (e.g. to add a bot).
				Name: "Discord Client ID",
				Regex: *regexp.MustCompile(`(?i:discord[\w.-]{0,20}?(?:client|app(?:lication)?)_?id)[\w.-]{0,10}` + assign + `\d{17,20}\b|` +
					`discord(?:app)?\.com(?:\\?/api(?:\\?/v\d{1,2})?)?\\?/oauth2\\?/authorize\?[^"'\s<>]{0,200}?client_id=\d{17,20}\b`),
				Sources: []string{"https://docs.discord.com/developers/topics/oauth2", "https://docs.discord.com/developers/reference"},
			},
			{
				// The ID of the user of a bot, the first part of its token in base64.
				Name:    "Discord Bot ID",
				Regex:   *regexp.MustCompile(keyed(`discord`, `bot[_-]?(?:token[_-]?)?id|token[_-]?id`, `\d{17,20}\b`)),
				Sources: []string{"https://docs.discord.com/developers/reference"},
			},
			{
				// A Docker ID (lowercase letters and digits, 4-30 characters) assigned to a
				// Docker username name (DOCKER_USER can be a user in the container), or in the
				// login to Docker Hub. It ends the value: dockerUser: user, is a variable.
				Name: "Docker Hub Username",
				Regex: *regexp.MustCompile(`(?:` + keyed(`docker`, `username|usr|hub[\w.-]{0,10}?user`, ``) +
					`|\bdocker login (?:-u|--username)[ =]["']?)[a-z0-9]{4,30}(?:["'\x60\s\\<&]|$)`),
				Sources: []string{"https://docs.docker.com/accounts/create-account/", "https://docs.docker.com/reference/cli/docker/login/"},
			},
			{
				// The app ID assigned to an app or client ID name, in the fb:app_id
				// meta tag, in FB.init, in the SDK URL or in the login dialog URL.
				Name: "Facebook Client ID",
				Regex: *regexp.MustCompile(`(?i)(?:facebook|fb)(?:[\w.-]{0,20}?|` + assign + `[\[{(][^\]})]{0,200}?|\.init\(\s{0,5}\{[^})]{0,200}?)` +
					`(?:app|client)_?id` + assign + `\d{13,17}\b|` +
					`fb:app_id["']?\s{1,5}content=["']?\d{13,17}\b|connect\.facebook\.net/[^\s"'<>]{0,100}?[?#&]appId=\d{13,17}\b|` +
					`facebook\.com/(?:v\d{1,3}\.\d{1,3}/)?dialog/oauth\?[^\s"'<>]{0,200}?\bclient_id=\d{13,17}\b`),
				Sources: []string{"https://developers.facebook.com/docs/sharing/webmasters"},
			},
			{
				Name: "Firebase Database",
				Regex: *regexp.MustCompile(`([a-z0-9.-]+\.firebaseio\.com|[a-z0-9.-]+\.firebaseapp\.com|` +
					`[a-z0-9-]+\.[a-z0-9-]+\.firebasedatabase\.app)`),
				Sources: []string{"https://firebase.google.com/docs/database/web/start"},
			},
			{
				// Registration tokens of the app instances.
				Name:  "Firebase Cloud Messaging Token",
				Regex: *regexp.MustCompile(`[\w-]{22}:APA91b[\w-]{100,}`),
			},
			{
				// Virtual-hosted (bucket.storage.googleapis.com) and path-style
				// (//storage.googleapis.com/bucket, also with JSON-escaped slashes, and
				// the JSON API) URLs, and gs:// URIs.
				Name: "Google Cloud Storage",
				Regex: *regexp.MustCompile(`\b[a-z0-9][a-z0-9._-]{1,220}\.storage\.googleapis\.com|(?:https?:)?\\?/\\?/` +
					`(?:storage\.googleapis\.com|commondatastorage\.googleapis\.com|storage\.cloud\.google\.com)\\?/` +
					`(?:(?:upload\\?/|download\\?/)?storage\\?/v1\\?/b\\?/)?[a-z0-9][a-z0-9._-]{2,220}|\bgs://[a-z0-9][a-z0-9._-]{2,220}`),
				FalsePositive: gcsFalsePositive,
				Sources: []string{
					"https://docs.cloud.google.com/storage/docs/request-endpoints",
					"https://docs.cloud.google.com/storage/docs/buckets#naming",
				},
			},
			{
				Name:    "Google OAuth Client ID",
				Regex:   *regexp.MustCompile(`\b\d{6,}-[0-9a-z]{32}\.apps\.googleusercontent\.com\b`),
				Sources: []string{"https://developers.google.com/identity/gsi/web/guides/get-google-api-clientid"},
			},
			{
				Name: "Google Calendar URI",
				Regex: *regexp.MustCompile(`https?:\/\/(?:calendar|www)\.google\.com` +
					`\/calendar\/(?:[0-9a-z.\/-]+\/)?embed\?src=[A-Za-z0-9%@&;=\-_\.\/]+`),
			},
			{
				// Client IDs of GitHub apps (Iv1., Iv23li...) and OAuth apps (Ov23li...,
				// or 20 hex digits in the authorization URL).
				Name: "GitHub OAuth Client ID",
				Regex: *regexp.MustCompile(`\b(?:Iv1\.[0-9a-f]{16}|[IO]v2\dli[A-Za-z0-9]{14})\b|` +
					`github\.com/login/oauth/authorize\?[^\s"'<>]{0,300}?\bclient_id=[0-9a-f]{20}\b`),
				Sources: []string{"https://docs.github.com/en/apps/oauth-apps/building-oauth-apps/authorizing-oauth-apps"},
			},
			{
				// 64 hex digits assigned to a GitLab app or client ID name, or in the
				// authorization URL.
				Name: "GitLab OAuth Application ID",
				Regex: *regexp.MustCompile(`(?i:gitlab[\w.-]{0,20}?(?:application|app|client)_?id)` + assign + `[0-9a-f]{64}\b|` +
					`/oauth/authorize\?[^\s"'<>]{0,300}?\bclient_id=[0-9a-f]{64}\b`),
				Sources: []string{"https://docs.gitlab.com/api/oauth2/"},
			},
			{
				// The stack of a customer and the endpoints of its region.
				Name:    "Grafana Cloud Instance",
				Regex:   *regexp.MustCompile(`\b[a-z0-9][a-z0-9-]{0,62}\.grafana\.net\b`),
				Sources: []string{"https://grafana.com/docs/grafana-cloud/security-and-account-management/authentication-and-permissions/"},
			},
			{
				// The clusters of the HashiCorp Cloud Platform, e.g. of Vault.
				Name:    "HashiCorp Cloud Cluster",
				Regex:   *regexp.MustCompile(`\b[a-z0-9][a-z0-9.-]{0,100}\.hashicorp\.cloud\b`),
				Sources: []string{"https://developer.hashicorp.com/vault/tutorials/get-started-hcp-vault-dedicated/vault-access-cluster"},
			},
			{
				// The RoleID of an AppRole (a UUID), assigned to a Vault or AppRole role ID
				// name: its SecretID is a secret.
				Name:    "HashiCorp Vault AppRole Role ID",
				Regex:   *regexp.MustCompile(keyed(`vault|approle`, `role[_-]?id`, uuid+`\b`)),
				Sources: []string{"https://developer.hashicorp.com/vault/docs/auth/approle"},
			},
			{
				Name:    "LinkedIn Client ID",
				Regex:   *regexp.MustCompile(`(?i:linkedin[\w.-]{0,20}?(?:client_?id|api_?key))[\w.-]{0,10}` + assign + `[a-z0-9]{12,14}\b`),
				Sources: []string{"https://learn.microsoft.com/en-us/linkedin/shared/authentication/client-credentials-flow"},
			},
			{
				// pubkey- and 32 hex digits (the format is not documented): the key of the
				// email validation in the browser.
				Name:    "Mailgun Public Validation Key",
				Regex:   *regexp.MustCompile(`\bpubkey-[0-9a-f]{32}\b`),
				Sources: []string{"https://documentation.mailgun.com/docs/mailgun/api-reference/authentication/"},
			},
			{
				// Public tokens (pk), for the browser.
				Name:    "Mapbox Public Token",
				Regex:   *regexp.MustCompile(`\bpk\.eyJ1Ijoi[\w-]{10,}\.[\w-]{20,}`),
				Sources: []string{"https://docs.mapbox.com/api/accounts/tokens/"},
			},
			{
				// The token of a project (32 hex digits, the format is not documented)
				// assigned to a Mixpanel token name or given to an SDK (mixpanel.init,
				// Mixpanel.initialize, sharedInstanceWithToken, MixpanelAPI.getInstance):
				// public by design, it only says which project the events go to.
				Name: "Mixpanel Project Token",
				Regex: *regexp.MustCompile(keyed(`mixpanel`, `token`, `[0-9a-f]{32}\b`) + `|` + configCall(`mixpanel`, `token`, `[0-9a-f]{32}\b`) +
					`|(?:(?i:\bmixpanel\.init\s{0,5}\(\s{0,5})|\bMixpanelAPI\.getInstance\([^()]{0,100}?|\bsharedInstanceWithToken:\s{0,5}@)["'\x60][0-9a-f]{32}\b`),
				Sources: []string{"https://docs.mixpanel.com/reference/project-token"},
			},
			{
				// 8 lowercase letters, assigned to an Atlas public key name, also in camelCase.
				// It works as the user of the private key, which is a secret. It ends the value:
				// atlasPublicKey:variable, is a variable.
				Name: "MongoDB Atlas Public API Key",
				Regex: *regexp.MustCompile(`(?:` + keyed(mongoAtlas, `pub(?:lic)?[\w-]{0,5}?key`, ``) +
					`|(?i:atlas(?:api)?pub(?:lic)?(?:api)?key)[\w.-]{0,10}` + assign +
					`|mongodbatlas"\s{0,5}\{[^}]{0,200}?public_key\s{0,5}=\s{0,5}"` +
					`)[a-z]{8}(?:["'\x60\s\\<&]|$)`),
				Sources: []string{"https://www.mongodb.com/docs/atlas/cli/current/atlas-cli-env-variables/", "https://developer.hashicorp.com/vault/docs/secrets/mongodbatlas"},
			},
			{
				// Assigned to a New Relic account ID name, or in the URLs of the Insights
				// APIs: the Insights keys need it.
				Name: "New Relic Account ID",
				Regex: *regexp.MustCompile(`(?i:new[_-]?relic[\w.-]{0,2}account[\w.-]{0,2}id)` + assign + `\d{3,10}\b|` +
					`\binsights-(?:api|collector)\.(?:eu\.)?newrelic\.com/v1/accounts/\d{3,10}\b`),
				Sources: []string{
					"https://docs.newrelic.com/docs/accounts/accounts-billing/account-structure/account-id/",
					"https://docs.newrelic.com/docs/data-apis/ingest-apis/event-api/introduction-event-api/",
				},
			},
			{
				// The key of the browser agent (NRJS-, or the older NRBR-): public by design.
				// The length is not documented.
				Name:    "New Relic Browser License Key",
				Regex:   *regexp.MustCompile(`\bNR(?:JS|BR)-[0-9a-f]{19}\b`),
				Sources: []string{"https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/"},
			},
			{
				// The token of a mobile app (42 hex digits, or the EU ones) with its -NRMA
				// suffix: it ships inside the app.
				Name:  "New Relic Mobile App Token",
				Regex: *regexp.MustCompile(`\b(?:AA[0-9a-f]{40}|eu01xx[0-9a-f]{36})-NRMA\b`),
				Sources: []string{
					"https://docs.newrelic.com/docs/mobile-monitoring/new-relic-mobile/maintenance/viewing-your-application-token/",
				},
			},
			{
				// The client ID of an app: 0oa and 17 characters, as every Okta ID has 20.
				Name:    "Okta Client ID",
				Regex:   *regexp.MustCompile(`\b0oa[A-Za-z0-9]{17}\b`),
				Sources: []string{"https://developer.okta.com/docs/guides/implement-grant-type/authcode/main/"},
			},
			{
				// A and 59 to 79 characters, in the JavaScript SDK URL or assigned to a
				// PayPal client ID name: it identifies an app, its secret is a secret.
				Name: "PayPal Client ID",
				Regex: *regexp.MustCompile(`paypal\.com/sdk/js\?[^\s"'<>]{0,300}?\bclient-id=A[\w-]{59,79}(?:[^\w-]|$)|` +
					`(?i:paypal[\w.-]{0,20}?client_?id)[\w.-]{0,10}` + assign + `A[\w-]{59,79}(?:[^\w-]|$)`),
				Sources: []string{
					"https://developer.paypal.com/api/rest/",
					"https://developer.paypal.com/sdk/js/configuration/",
				},
			},
			{
				// 24 hex digits assigned to a Plaid client ID name: its secret is a secret.
				Name:    "Plaid Client ID",
				Regex:   *regexp.MustCompile(keyed(`plaid`, `client[_-]?id`, `[0-9a-f]{24}\b`)),
				Sources: []string{"https://plaid.com/docs/quickstart/glossary/"},
			},
			{
				// The client ID of a connected app. Only the prefix is known, the format is not
				// documented. Its secret is a secret.
				Name:    "Salesforce Consumer Key",
				Regex:   *regexp.MustCompile(`\b3MVG9[\w.+/=]{80,251}`),
				Sources: []string{"https://docs.aws.amazon.com/secretsmanager/latest/userguide/mes-partner-salesforce.html"},
			},
			{
				// The My Domain of an org, also of a sandbox or a scratch org (in lowercase:
				// the documentation writes MyDomainName).
				Name:    "Salesforce Instance",
				Regex:   *regexp.MustCompile(`\b[a-z0-9][a-z0-9.-]{0,80}\.my\.salesforce\.com\b`),
				Sources: []string{"https://help.salesforce.com/s/articleView?id=sf.domain_name_url_format_changes_no_enhanced.htm&language=en_US&type=5"},
			},
			{
				// The public key, the host and the project ID, also with JSON-escaped
				// slashes: public by design. A DSN with a secret key is found as
				// credentials in a URL.
				Name:    "Sentry DSN",
				Regex:   *regexp.MustCompile(`https?:\\?/\\?/[0-9a-f]{32}@[A-Za-z0-9][A-Za-z0-9.-]{0,251}(?::\d{1,5})?\\?/\d{1,12}\b`),
				Sources: []string{"https://docs.sentry.io/security-legal-pii/security/token-formats/"},
			},
			{
				// The client ID of an app (32 hex digits), assigned to a Shopify client ID or
				// API key name (SHOPIFY_API_KEY): its secret is a secret.
				Name:    "Shopify App Client ID",
				Regex:   *regexp.MustCompile(`(?i:shopify[\w.-]{0,20}?(?:client_?id|api_?key))[\w.-]{0,10}` + assign + `[0-9a-f]{32}\b`),
				Sources: []string{"https://shopify.dev/docs/apps/build/cli-for-apps/app-configuration"},
			},
			{
				Name:    "Shopify Store Domain",
				Regex:   *regexp.MustCompile(`\b[A-Za-z0-9][A-Za-z0-9-]{0,62}\.myshopify\.com\b`),
				Sources: []string{"https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens/offline-access-tokens"},
			},
			{
				// Assigned to a Snowflake account name (SNOWFLAKE_ACCOUNT), also in an object or
				// to account in a connection call: it is in the URL of the account.
				Name: "Snowflake Account Identifier",
				Regex: *regexp.MustCompile(keyed(`snowflake`, `account`, snowflakeAccount) + `|` +
					configCall(`snowflake`, `account`, snowflakeAccount)),
				Sources: []string{"https://docs.snowflake.com/en/user-guide/admin-account-identifier"},
			},
			{
				// Assigned to sonar.host.url or SONAR_HOST_URL: the server analyzing the
				// project. Local and documentation hosts are skipped.
				Name: "SonarQube Server URL",
				Regex: *regexp.MustCompile(`(?i:sonar[._-]?host[._-]?url)` + assign +
					`https?://[A-Za-z0-9](?:[A-Za-z0-9.-]{0,251}[A-Za-z0-9])?(?::\d{1,5})?(?:/[\w./~%-]{0,200})?`),
				FalsePositive: urlHostFalsePositive,
				Sources: []string{
					"https://docs.sonarsource.com/sonarqube-server/analyzing-source-code/analysis-parameters/parameters-not-settable-in-ui",
				},
			},
			{
				// The application ID (also of the Sandbox): public, it identifies the app
				// in the browser. The length is not documented.
				Name:    "Square Application ID",
				Regex:   *regexp.MustCompile(`\b(?:sandbox-)?sq0id[bp]-[\w-]{20,22}(?:[^\w-]|$)`),
				Sources: []string{"https://developer.squareup.com/docs/oauth-api/overview"},
			},
			{
				// Public by design: it identifies the account in the browser.
				Name:    "Stripe Publishable Key",
				Regex:   *regexp.MustCompile(`\bpk_(?:live|test)_[0-9a-zA-Z]{24,247}\b`),
				Sources: []string{"https://docs.stripe.com/keys"},
			},
			{
				// Retrievable with the publishable key: it completes a single payment.
				Name:    "Stripe PaymentIntent Client Secret",
				Regex:   *regexp.MustCompile(`\bpi_[0-9a-zA-Z]{14,}_secret_[0-9a-zA-Z]{14,}\b`),
				Sources: []string{"https://docs.stripe.com/api/payment_intents/object"},
			},
			{
				// The host of the API of a project: its reference has 20 characters.
				Name:    "Supabase Project URL",
				Regex:   *regexp.MustCompile(`\b[a-z0-9]{20}\.supabase\.co\b`),
				Sources: []string{"https://supabase.com/docs/guides/api"},
			},
			{
				// sb_publishable_ and 31 characters (the length is not documented): public
				// by design, the secret keys are secrets.
				Name:    "Supabase Publishable Key",
				Regex:   *regexp.MustCompile(`\bsb_publishable_[\w-]{31}(?:[^\w-]|$)`),
				Sources: []string{"https://supabase.com/docs/guides/api/api-keys"},
			},
			{
				// 32 hex digits assigned to a Trello key name or in the key parameter of an API
				// URL (also of client.js): public by design, the token is a secret.
				Name: "Trello API Key",
				Regex: *regexp.MustCompile(keyed(`trello`, `key`, `[0-9a-f]{32}\b`) +
					`|\btrello\.com/1/[^\s"'<>]{0,200}?[?&]key=[0-9a-f]{32}\b`),
				Sources: []string{"https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/"},
			},
			{
				// Account (AC) and API key (SK) SIDs: their secrets are secrets.
				Name:    "Twilio SID",
				Regex:   *regexp.MustCompile(`\b(?:AC|SK)[0-9a-f]{32}\b`),
				Sources: []string{"https://www.twilio.com/docs/glossary/what-is-a-sid"},
			},
			{
				// API keys (consumer keys) and OAuth 2.0 client IDs.
				Name:  "Twitter Client ID",
				Regex: *regexp.MustCompile(`(?i:twitter[\w.-]{0,20}?(?:client_?id|consumer_?key|api_?key))[\w.-]{0,10}` + assign + `[\w-]{18,34}\b`),
			},
			/*
				TOO MANY FALSE POSITIVES
				{
					"BTC address",
					`([13]|bc1)[A-HJ-NP-Za-km-z1-9]{27,34}`,
				},
			*/
			/*
				HOW TO AVOID VERY VERY LONG BASE64 IMAGES ???
				{
					"Base64-encoded JSON",
					`ey(A|B)[A-Za-z0-9+\/]{20,}(={0,2})`,
				},
			*/
		}
	})

	return infos
}

// GetInfoScanner returns the scanner for the predefined info regexes.
func GetInfoScanner() *InfoScanner {
	onceInfoScanner.Do(func() {
		infoScanner = NewInfoScanner(GetInfoRegexes())
	})

	return infoScanner
}

// NewInfoScanner returns a scanner for the given infos.
func NewInfoScanner(infos []Info) *InfoScanner {
	s := &InfoScanner{infos: slices.Clone(infos)}
	regexes := make([]*regexp.Regexp, len(s.infos))

	for i := range s.infos {
		regexes[i] = &s.infos[i].Regex
	}

	s.engine = NewEngine(regexes)

	return s
}

// Scan returns the infos found in body, in the order of the infos list,
// excluding the false positives.
func (s *InfoScanner) Scan(target, body string) []InfoMatched {
	var results []InfoMatched

	for i, locs := range s.engine.FindAllIndex(body) {
		info := &s.infos[i]

		for _, loc := range locs {
			match := body[loc[0]:loc[1]]
			if containsFalsePositive(match, info.FalsePositives) ||
				info.FalsePositive != nil && info.FalsePositive(body, loc[0], loc[1]) {
				continue
			}

			results = append(results, InfoMatched{Info: *info, URL: target, Match: match})
		}
	}

	return results
}

// urlHostFalsePositive reports whether the URL in the match text[start:end]
// (name = URL) has a local host or one reserved for documentation.
func urlHostFalsePositive(text string, start, end int) bool {
	match := text[start:end]
	host := uriHost(match[strings.Index(match, "://")+len("://"):])

	return exampleHost(host) || localHost(host)
}

// ipv4FalsePositive reports whether text[start:end] is part of a longer
// word or number, e.g. 1.2.3.4 in v1.2.3.4 or 11.2.3.4, or of a longer
// dotted number, e.g. in the version 1.2.3.4.5 or the OID 1.3.6.1.4.1
// (the dot of a sentence end is not part of it). An escape sequence is not
// part of a word, e.g. \n10.0.0.1 or \u003e10.0.0.1 in a JSON string.
func ipv4FalsePositive(text string, start, end int) bool {
	// at returns the byte at i, 0 if i is out of text.
	at := func(i int) byte {
		if i < 0 || i >= len(text) {
			return 0
		}

		return text[i]
	}
	isDigit := func(c byte) bool { return '0' <= c && c <= '9' }
	isWord := func(c byte) bool { return isDigit(c) || c == '_' || 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' }
	before, after := start-1, end

	return (isWord(at(before)) && !escapeSuffix.MatchString(text[max(0, start-len(`\u0000`)):start])) ||
		(at(before) == '.' && isDigit(at(before-1))) ||
		(at(after) == '.' && isDigit(at(after+1)))
}

// emailFalsePositive reports whether the match text[start:end] is not an
// email address: its top-level domain is not one (e.g. logo@2x.png), it
// starts inside an escape sequence (e.g. \u003cuser@example.com), or it is the
// user and host of a URL (e.g. https://user:password@example.com, also with
// escaped slashes), of an SSH location (e.g. git@github.com:org/repo) or of an
// Oracle JDBC URL (e.g. jdbc:oracle:thin:user/password@host).
func emailFalsePositive(text string, start, end int) bool {
	match := text[start:end]
	tld := strings.ToLower(match[strings.LastIndexByte(match, '.')+1:])

	// With a label before it: some TLDs are listed as wildcards (*.np).
	if _, icann := publicsuffix.PublicSuffix("x." + tld); !icann {
		return true
	}

	if start > 0 && text[start-1] == '\\' {
		return true
	}

	return inURLUserinfo(text, start) || inOracleURL(text, start) || sshPathPrefix.MatchString(text[end:])
}

// gitSSHFalsePositive reports whether the match text[start:end] is the part of
// a URL of another scheme after a colon in its userinfo, not an SSH location
// (e.g. s3cret@pg.shop.io:3306/app in mysql://app:s3cret@pg.shop.io:3306/app).
func gitSSHFalsePositive(text string, start, _ int) bool {
	return inURLUserinfo(text, start) && !strings.Contains(text[max(0, start-len("ssh+git://")):start], "ssh")
}

// inURLUserinfo reports whether text[start:] starts in the userinfo of a URL,
// i.e. after // and the bytes of a userinfo (e.g. https://user:), also with
// escaped slashes (\/\/), as in JSON.
func inURLUserinfo(text string, start int) bool {
	i := start
	for i > 0 && start-i < maxUserinfo && strings.IndexByte(userinfoChars, text[i-1]) >= 0 {
		i--
	}

	before := text[max(0, i-len(`\/\/`)):i]

	return strings.HasSuffix(before, "//") || strings.HasSuffix(before, `\/\/`)
}

// inOracleURL reports whether text[start:] starts in the password of an Oracle
// JDBC URL (jdbc:oracle:thin:user/password@host), where the user and the
// password come before the host without //.
func inOracleURL(text string, start int) bool {
	i := start
	for i > 0 && start-i < maxUserinfo && (text[i-1] == '/' || strings.IndexByte(userinfoChars, text[i-1]) >= 0) {
		i--
	}

	return strings.Contains(strings.ToLower(text[i:start]), "jdbc:oracle:")
}

// gcsFalsePositive reports whether the bucket of the match text[start:end] is
// the first segment of a JSON API path instead, as in /storage/v1/... or
// /upload/storage/v1/... (the bucket, e.g. a placeholder, did not match).
func gcsFalsePositive(text string, _, end int) bool {
	rest := strings.ReplaceAll(text[end:min(end+len(`\/storage\/v1\/`), len(text))], `\/`, "/")

	return strings.HasPrefix(rest, "/storage/v1") ||
		strings.HasPrefix(rest, "/v1") && strings.HasSuffix(text[:end], "/storage")
}

// RemoveDuplicateInfos removes duplicates from Infos found.
func RemoveDuplicateInfos(input []InfoMatched) []InfoMatched {
	keys := make(map[string]bool)
	list := []InfoMatched{}

	for _, entry := range input {
		if _, value := keys[entry.Match]; !value {
			keys[entry.Match] = true
			list = append(list, entry)
		}
	}

	return list
}
