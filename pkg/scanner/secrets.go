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
	"net/netip"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// Secret struct.
// Name = the name that identifies the secret.
// Description.
// Regex = The regular expression matching the secret.
// FalsePositives = A list of known false positives.
// PoC = cli command to check if the secret is valid or not.
// Sources = links to the official documentation of the format, if any.
type Secret struct {
	Name           string
	Description    string
	Regex          regexp.Regexp
	FalsePositives []string
	Poc            string
	Sources        []string
}

// SecretMatched struct.
// Secret = The secret matched (struct).
// Url = url in which is present the secret.
// Match = the string matching the regex.
type SecretMatched struct {
	Secret Secret
	URL    string
	Match  string
}

// SecretScanner hunts for secrets in response bodies (see Engine).
// It is safe for concurrent use.
type SecretScanner struct {
	secrets []Secret
	engine  *Engine
}

// assign is the part of a regex between a name and the value assigned to it,
// e.g. name = "value", "name": "value", name => 'value' or name := `value`.
const assign = `["']?\s{0,5}[:=]{1,2}>?\s{0,5}["'\x60]?`

// secretWords are words naming a secret, e.g. in api_key, authToken,
// client-secret, PASSWORD, pwd or credentials.
const secretWords = `key|token|secret|pass|pwd|auth|cred`

// keyed returns a regex for a value assigned to a name containing one of
// the vendors and then one of the words, e.g. ZENDESK_API_TOKEN = "value" or
// Zendesk:ApiToken (any name containing a vendor if words is empty), or to
// a name containing one of the words (secretWords if empty) in an object,
// a YAML block or an INI section of a vendor, e.g. "zendesk": {"token":
// "value"}, zendesk:\n  token: value or [zendesk]\ntoken = value.
func keyed(vendors, words, value string) string {
	name := `(?i:(?:` + vendors + `)[\w.:-]{0,30})`
	inner := secretWords

	if words != "" {
		name = `(?i:(?:` + vendors + `)[\w.:-]{0,30}?(?:` + words + `))[\w.-]{0,10}`
		inner = words
	}

	// The vendor names an object, a list, a YAML block or an INI section.
	nested := `(?i:(?:` + vendors + `)(?:[\w.-]{0,20}["']?\s{0,5}[:=]{1,2}>?[ \t]{0,5}(?:[\[{(]|\r?\n)|[\w. "'-]{0,20}\][ \t]{0,5}\r?\n)` +
		`[^\]})]{0,200}?(?:` + inner + `))[\w.-]{0,10}`

	return `(?:` + name + `|` + nested + `)` + assign + value
}

// configCall returns a regex for a value assigned to one of the words in the
// arguments of a call on one of the vendors, e.g. cloudinary.config({cloud_name:
// "value"}) or snowflake.connector.connect(account="value"), or given as the
// first argument (quoted) if words is empty, e.g. cohere.ClientV2("value"). The
// arguments before the value can't hold another call.
func configCall(vendors, words, value string) string {
	if words == "" {
		return `(?i:(?:` + vendors + `)[\w.]{0,30}\s{0,5}\(\s{0,5})["'\x60]` + value
	}

	return `(?i:(?:` + vendors + `)[\w.]{0,30}\s{0,5}\([^()]{0,300}?\b(?:` + words + `))` + assign + value
}

// atlassianDCToken is a personal or HTTP access token of an Atlassian Data
// Center product: an ID of 12 digits, ":" and 20 random bytes, in base64 (the
// digits encode in groups of 3, ":" and the next byte as O and g to v).
const atlassianDCToken = `(?:[MNO][DTjz][AEIMQUYcgk][w-z0-5]){4}O[g-v][A-Za-z0-9+/]{26}(?:[^A-Za-z0-9+/=]|$)`

const (
	// connUser and connPassword are the user and the password of a connection URI,
	// percent-encoded as RFC 3986 requires. They leave out placeholders such as %s,
	// ${PASSWORD}, <password>, [password], {0}, *** and ... (a password can't start
	// with $ or a dot).
	connUser     = `(?:[\w.~!+,=-]|%[0-9A-Fa-f]{2}){1,64}`
	connPassword = `(?:[\w~!+,=-]|%[0-9A-Fa-f]{2})(?:[\w.~!$+,=-]|%[0-9A-Fa-f]{2}){2,127}`
	// connHost is the first host of a connection URI, with its port: a name or an
	// address, or an IPv6 address in brackets. A name can't start with an uppercase
	// letter (HOST, DB_HOST) and what follows can't continue it (host:port, db_host,
	// host-${ENV}, tcp(host:port)).
	connHost = `(?:[a-z0-9](?:[A-Za-z0-9.-]{0,251}[A-Za-z0-9])?|\[[0-9A-Fa-f:.]{2,45}\])(?::\d{1,5})?(?:[^\w.:$%{(\[-]|$)`
	// jsonSpace is the space between JSON tokens, also escaped (JSON in a string).
	jsonSpace = `(?:\s|\\{1,8}[nrt]){0,20}`
	// mongoAtlas are the names of the MongoDB Atlas API keys, as in
	// MONGODB_ATLAS_PRIVATE_API_KEY or MCLI_PRIVATE_API_KEY (not atlassian).
	mongoAtlas = `mongodb|mcli_|atlas(?:[_.:-]|\b)`
	// uuid is a UUID in lowercase.
	uuid = `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`
	// propPassword is the password of a connection string made of properties
	// (key=value). It leaves out placeholders such as ${PASSWORD}, {0}, <password>,
	// [password], %s and *** (a password can't start with $, % or *).
	propPassword = `[^\s;&"'\x60<>{}\[\]$%*][^\s;&"'\x60<>{}\[\]]{2,127}`
	// propPairs are the properties between two others of a connection string
	// (key=value;), on the same line: the keys are words, as in Initial Catalog
	// or Trusted_Connection.
	propPairs = `[ \t]{0,3}(?:[A-Za-z][A-Za-z _]{0,40}=[^;"'\n<>]{0,100};[ \t]{0,3}){0,8}?`
	// propServer is the property of a connection string naming its server. The
	// space around = is \s on purpose, here and after the password name: with
	// [ \t], the engine would expand it and look for every "host=" and "server="
	// instead of the rarer "password=".
	propServer = `(?i:Data Source|Server|Host)\s{0,3}=\s{0,3}[^;"'\s<>=]{1,128}`
	// propTrusted is the property of a connection string asking for the account
	// of the process instead of a user and a password.
	propTrusted = `(?i:Integrated Security|Trusted_Connection)\s{0,3}=\s{0,3}(?i:SSPI|true|yes)\b`
	// propBraced is a password in braces, as ODBC allows for special characters
	// such as ;. One of them is required, and quotes, brackets and $ are not
	// allowed: placeholders of templates and format strings ({0}, {password},
	// {db.password}, {{pwd}}) and code ({os.getenv("PWD")}) are left out.
	propBraced = `\{[^\s{}"'\x60<>\[\]()$%*]{0,127}?[^\w.\s{}"'\x60<>\[\]()$%*-][^\s{}"'\x60<>\[\]()$%*]{0,127}\}`
	// kubeBase64 is a value in base64 of a Kubernetes Secret: a multiple of 4
	// characters, from 8 to 1004. A longer value (a certificate) is not looked for.
	kubeBase64 = `(?:[A-Za-z0-9+/]{4}){2,250}(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?`
	// kubeData is the first entry of the data of a Kubernetes Secret in YAML, a
	// key and a value in base64, without the end of its line.
	kubeData = `[ \t]{1,20}["']?[\w.-]{1,253}["']?:[ \t]{0,5}["']?` + kubeBase64
	// kubeJSONData is the data of a Kubernetes Secret in JSON (also in a string),
	// up to its first value in base64, without the closing quote.
	kubeJSONData = `\\?"data\\?"` + jsonSpace + `:` + jsonSpace + `\{` + jsonSpace + `\\?"[\w.-]{1,253}\\?"` + jsonSpace + `:` + jsonSpace +
		`\\?"` + kubeBase64
)

// connectionString returns a regex for a connection URI of the given schemes,
// in any case, with a user, a password and a host (see connFalsePositive).
func connectionString(schemes string) string {
	return `(?i:` + schemes + `):\\?/\\?/` + connUser + `:` + connPassword + `@` + connHost
}

// documentationAddresses are the address blocks reserved for documentation.
var documentationAddresses = []netip.Prefix{ //nolint: gochecknoglobals
	netip.MustParsePrefix("192.0.2.0/24"),
	netip.MustParsePrefix("198.51.100.0/24"),
	netip.MustParsePrefix("203.0.113.0/24"),
	netip.MustParsePrefix("2001:db8::/32"),
}

var (
	secrets           []Secret       //nolint: gochecknoglobals
	onceSecret        sync.Once      //nolint: gochecknoglobals
	secretScanner     *SecretScanner //nolint: gochecknoglobals
	onceSecretScanner sync.Once      //nolint: gochecknoglobals
)

// GetSecretRegexes returns a slice of all the secret structs.
func GetSecretRegexes() []Secret {
	onceSecret.Do(func() {
		secrets = []Secret{
			{
				// The Secret Key of an account: the version (A3), the account ID (6
				// characters) and the key, 34 letters and digits in groups.
				"1Password Secret Key",
				"1Password Secret Key",
				*regexp.MustCompile(`\bA3-[A-Z0-9]{6}-(?:[A-Z0-9]{11}|[A-Z0-9]{6}-[A-Z0-9]{5})-[A-Z0-9]{5}-[A-Z0-9]{5}-[A-Z0-9]{5}\b`),
				[]string{},
				"?",
				[]string{"https://support.1password.com/secret-key-security/"},
			},
			{
				// ops_ and the credentials of the service account (JSON in base64). A token
				// of 1000 characters or more has no end check: regexes are limited to that.
				"1Password Service Account Token",
				"1Password Service Account Token",
				*regexp.MustCompile(`\bops_eyJ[\w+/-]{250,1000}={0,2}`),
				[]string{},
				"?",
				[]string{"https://www.1password.dev/service-accounts/security"},
			},
			{
				// Keys (also with the aio_ prefix) assigned to an Adafruit name or to an AIO
				// key name, or with the prefix alone.
				"Adafruit API Key",
				"Adafruit API Key",
				*regexp.MustCompile(keyed(`adafruit`, ``, `[A-Za-z0-9_]{32}\b`) + `|(?i:(?:\b|_)aio[_-]?(?:key|token))[\w.-]{0,10}` + assign +
					`[A-Za-z0-9_]{32}\b|\baio_[A-Za-z0-9]{28}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				"Adobe Client Secret",
				"Adobe Client Secret",
				*regexp.MustCompile(`\bp8e-[\w-]{32}`),
				[]string{},
				"?",
				[]string{
					"https://experienceleague.adobe.com/en/docs/experience-manager-learn/cloud-service/developing/extensibility/app-builder/server-to-server-auth",
				},
			},
			{
				// Personal access tokens: pat and the token ID, then the secret.
				"Airtable API Key",
				"Airtable API Key",
				*regexp.MustCompile(`\bpat[A-Za-z0-9]{14}\.[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://airtable.com/developers/web/guides/personal-access-tokens"},
			},
			{
				// Assigned to an admin, write or secret key name, or to an Algolia API key
				// name: search-only keys, which look the same, are public (e.g. the apiKey
				// in the algolia object of the site search widgets).
				"Algolia API Key",
				"Algolia API Key",
				*regexp.MustCompile(keyed(`algolia`, `(?:admin|write|secret)[\w.-]{0,5}?key`, `[0-9a-f]{32}\b`) +
					`|(?i:algolia[\w.:-]{0,30}?api[\w.-]{0,5}?key)[\w.-]{0,10}` + assign + `[0-9a-f]{32}\b`),
				[]string{},
				"?",
				[]string{"https://www.algolia.com/doc/guides/security/api-keys/"},
			},
			{
				"Alibaba Access Key ID",
				"Alibaba Access Key ID",
				*regexp.MustCompile(`\bLTAI[A-Za-z0-9]{12,20}\b`),
				[]string{},
				"?",
				[]string{"https://www.alibabacloud.com/help/en/ram/developer-reference/api-ims-2019-08-15-listaccesskeys"},
			},
			{
				// Assigned to an Alibaba Cloud secret or key name (bucket names can look the
				// same), or to AccessKeySecret.
				"Alibaba Secret Key",
				"Alibaba Secret Key",
				*regexp.MustCompile(keyed(`alibaba|aliyun|alicloud`, `secret|key`, `[A-Za-z0-9]{30}\b`) + `|(?i:access_?key_?secret)[\w.-]{0,10}` +
					assign + `[A-Za-z0-9]{30}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Personal, service account, workspace and Enterprise keys: 95
				// characters, like Admin API keys.
				"Anthropic API Key",
				"Anthropic API Key",
				*regexp.MustCompile(`\bsk-ant-api\d{2}-[\w-]{95}(?:\b|[^\w-]|$)`),
				[]string{},
				"?",
				[]string{"https://platform.claude.com/docs/en/api/admin/api_keys/list"},
			},
			{
				"Anthropic Admin API Key",
				"Anthropic Admin API Key",
				*regexp.MustCompile(`\bsk-ant-admin\d{2}-[\w-]{95}(?:\b|[^\w-]|$)`),
				[]string{},
				"?",
				[]string{"https://platform.claude.com/docs/en/manage-claude/admin-api"},
			},
			{
				// OAuth access (oat) and refresh (ort) tokens, e.g. of Claude Code or
				// of workload identity federation.
				"Anthropic OAuth Token",
				"Anthropic OAuth Token",
				*regexp.MustCompile(`\bsk-ant-(?:oat|ort)\d{2}-[\w-]{95}(?:\b|[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://platform.claude.com/docs/en/cli-sdks-libraries/cli/authentication",
					"https://platform.claude.com/docs/en/manage-claude/workload-identity-federation",
				},
			},
			{
				// The sessionKey cookie of claude.ai: the format is not documented, the
				// length is the one of the other keys.
				"Anthropic Session Key",
				"Anthropic Session Key",
				*regexp.MustCompile(`\bsk-ant-sid\d{2}-[\w-]{95}(?:\b|[^\w-]|$)`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Assigned to any Asana name: the IDs are numbers.
				"Asana Client Secret",
				"Asana Client Secret",
				*regexp.MustCompile(keyed(`asana`, ``, `[A-Za-z0-9]{32}\b`)),
				[]string{},
				"?",
				[]string{},
			},
			{
				// A version, the IDs of the user and of the token, and the token (32
				// characters or more), or the older 0/ and 32 hex digits, assigned to any
				// Asana name. The format is not documented: tokens are opaque.
				"Asana Personal Access Token",
				"Asana Personal Access Token",
				*regexp.MustCompile(keyed(`asana`, ``, `(?:\d/\d{1,20}(?:/\d{1,20})?:[A-Za-z0-9]{32,128}|0/[0-9a-f]{32})\b`)),
				[]string{},
				"?",
				[]string{"https://developers.asana.com/docs/personal-access-token"},
			},
			{
				// Tokens created before 2023 (24 characters) assigned to a token name: the
				// account IDs can look the same.
				"Atlassian API Token",
				"Atlassian API Token",
				*regexp.MustCompile(keyed(`atlassian|confluence|jira`, `token|key|secret|password`, `[A-Za-z0-9]{24}\b`)),
				[]string{},
				"?",
				[]string{"https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/"},
			},
			{
				// API tokens (ATATT) and access tokens (ATCTT): = and 8 hex digits at the end.
				"Atlassian Cloud API Token",
				"Atlassian Cloud API Token",
				*regexp.MustCompile(`\bAT(?:ATT|CTT)[\w-]{20,250}=[0-9A-F]{8}\b`),
				[]string{},
				"?",
				[]string{
					"https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/",
					"https://community.atlassian.com/forums/Bitbucket-questions/Can-we-confirm-BitBucket-s-token-prefixes/qaa-p/3096469",
				},
			},
			{
				// Bitbucket HTTP access tokens (BBDC-), and Jira, Confluence, Bitbucket and
				// Bamboo personal access tokens after Bearer or within 200 characters after a
				// product name.
				"Atlassian Data Center Access Token",
				"Atlassian Data Center Access Token",
				*regexp.MustCompile(`\bBBDC-` + atlassianDCToken + `|(?i:\bbearer\s{1,5}|(?:atlassian|confluence|jira|bitbucket|bamboo)` +
					`[\s\S]{0,200}?\b)` + atlassianDCToken),
				[]string{},
				"?",
				[]string{
					"https://confluence.atlassian.com/enterprise/using-personal-access-tokens-1026032365.html",
					"https://confluence.atlassian.com/bitbucketserver/http-access-tokens-939515499.html",
					"https://support.atlassian.com/bitbucket-data-center/kb/git-commands-failing-with-url-rejected-when-using-http-access-tokens-as-basic-auth/",
				},
			},
			{
				// 64 characters assigned to an Auth0 secret name (client secrets and the
				// session secret of the SDKs): the telemetry of the SDKs (auth0Client, in
				// base64) can look the same. The client IDs are info.
				"Auth0 Client Secret",
				"Auth0 Client Secret",
				*regexp.MustCompile(keyed(`auth0`, `secret`, `[\w-]{64}(?:[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{"https://auth0.com/docs/get-started/applications/application-settings"},
			},
			{
				// Long-term (AKIA) and temporary (ASIA) access key IDs, STS bearer tokens (ABIA)
				// and context-specific credentials (ACCA): the other unique ID prefixes
				// identify users, roles, groups and policies.
				"AWS Access Key",
				"AWS Access Key",
				*regexp.MustCompile(`\b(?:AKIA|ASIA|ABIA|ACCA)[A-Z0-9]{16}\b`),
				[]string{},
				"?",
				[]string{"https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html"},
			},
			{
				// Long-term keys.
				"AWS Bedrock API Key",
				"AWS Bedrock API Key",
				*regexp.MustCompile(`\bABSK[A-Za-z0-9+/]{109,269}={0,2}`),
				[]string{},
				"?",
				[]string{"https://docs.aws.amazon.com/bedrock/latest/userguide/api-keys.html"},
			},
			{
				// Assigned to a secret access key name, or right after its access key ID
				// (e.g. the credentials CSV).
				"AWS Secret Key",
				"AWS Secret Key",
				*regexp.MustCompile(`(?i:aws_?secret_?(?:access_?)?key|secret_?access_?key)` + assign +
					`[A-Za-z0-9/+]{40}(?:[^A-Za-z0-9/+=]|$)|\b(?:AKIA|ASIA)[A-Z0-9]{16}["',:;|\s]{1,5}[A-Za-z0-9/+]{40}(?:[^A-Za-z0-9/+=]|$)`),
				[]string{},
				"?",
				[]string{
					"https://docs.aws.amazon.com/IAM/latest/UserGuide/id_credentials_access-keys.html",
					"https://docs.aws.amazon.com/sdkref/latest/guide/file-format.html",
					"https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html",
				},
			},
			{
				// Temporary credentials: the base64 starts with the same bytes.
				"AWS Session Token",
				"AWS Session Token",
				*regexp.MustCompile(`\b(?:IQoJb3JpZ2luX2Vj|FwoGZXIvYXdz|FQoGZXIvYXdz)[A-Za-z0-9+/]{100,}={0,2}`),
				[]string{},
				"?",
				[]string{},
			},
			{
				"AWS Bedrock Short-term API Key",
				"AWS Bedrock Short-term API Key",
				*regexp.MustCompile(`bedrock-api-key-[A-Za-z0-9+/]{100,}={0,2}`),
				[]string{},
				"?",
				[]string{"https://github.com/aws/aws-bedrock-token-generator-python"},
			},
			{
				"AWS MWS Key",
				"AWS MWS Key",
				*regexp.MustCompile(`\bamzn\.mws\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Azure identifiable keys: a fixed signature at a fixed position.
				"Azure Storage Account Key",
				"Azure Storage Account Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{76}\+ASt[A-Za-z0-9+/]{5}[AQgw]==`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-storage-account-access-key",
				},
			},
			{
				"Azure Cosmos DB Key",
				"Azure Cosmos DB Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{76}ACDb[A-Za-z0-9+/]{5}[AQgw]==`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-cosmos-db-account-access-key",
				},
			},
			{
				"Azure Batch Key",
				"Azure Batch Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{76}\+ABa[A-Za-z0-9+/]{5}[AQgw]==`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
				},
			},
			{
				"Azure API Management Key",
				"Azure API Management Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{76}APIM[A-Za-z0-9+/]{5}[AQgw]==`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
				},
			},
			{
				// Identifier, expiry and signature, for the management API and for the
				// configuration repository (git), also with the & escaped in HTML.
				"Azure API Management Access Token",
				"Azure API Management Access Token",
				*regexp.MustCompile(`SharedAccessSignature (?:uid=)?[\w-]{1,64}&(?:amp;)?(?:ex=)?[\dT:.Z-]{12,28}&(?:amp;)?` +
					`(?:sn=)?[A-Za-z0-9+/]{86}==|\bgit&(?:amp;)?\d{12}&(?:amp;)?[A-Za-z0-9+/]{86}==`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/rest/api/apimanagement/apimanagementrest/azure-api-management-rest-api-authentication",
				},
			},
			{
				// Identifiable keys, or keys in the code parameter of function URLs.
				"Azure Function Key",
				"Azure Function Key",
				*regexp.MustCompile(`[\w-]{44}AzFu[\w-]{5}[AQgw]==|\.azurewebsites\.net/api/[\w/.-]{0,200}\?[^\s"'<>]{0,300}?` +
					`\bcode=[\w%/+-]{40,200}={0,2}`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-function-master-api-key",
					"https://learn.microsoft.com/en-us/azure/azure-functions/function-keys-how-to",
				},
			},
			{
				"Azure AI Search Key",
				"Azure AI Search Key",
				*regexp.MustCompile(`[A-Za-z0-9]{42}AzSe[A-D][A-Za-z0-9]{5}`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
				},
			},
			{
				"Azure Container Registry Key",
				"Azure Container Registry Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{42}\+ACR[A-D][A-Za-z0-9+/]{5}`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
				},
			},
			{
				// Service Bus, Event Hubs, Relay, IoT and Event Grid identifiable keys.
				"Azure Shared Access Key",
				"Azure Shared Access Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{33}(?:\+ASb|\+AEh|\+ARm|AIoT|AZEG)[A-P][A-Za-z0-9+/]{5}=`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
				},
			},
			{
				"Azure Cache for Redis Key",
				"Azure Cache for Redis Key",
				*regexp.MustCompile(`[A-Za-z0-9+/]{33}AzCa[A-P][A-Za-z0-9+/]{5}=`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
				},
			},
			{
				// Common annotated security keys: 84 characters, JQQJ at position 52
				// and the service signature at position 76.
				"Azure DevOps Personal Access Token",
				"Azure DevOps Personal Access Token",
				*regexp.MustCompile(`[A-Za-z0-9]{52}JQQJ9[9DH][A-Za-z0-9][A-L][A-Za-z0-9]{16}AZDO[A-Za-z0-9]{4}`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/azure/devops/organizations/accounts/use-personal-access-tokens-to-authenticate",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-devops-personal-access-token",
					"https://devblogs.microsoft.com/engineering-at-microsoft/common-annotated-security-keys/",
				},
			},
			{
				// OpenAI, Speech, Translator, Vision and the other AI services.
				"Azure AI Services Key",
				"Azure AI Services Key",
				*regexp.MustCompile(`[A-Za-z0-9]{52}JQQJ9[9DH][A-Za-z0-9][A-L][A-Za-z0-9]{16}ACOG[A-Za-z0-9]{4}`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
					"https://devblogs.microsoft.com/engineering-at-microsoft/common-annotated-security-keys/",
				},
			},
			{
				// App Configuration, Event Grid, Fluid Relay, Mixed Reality, Maps,
				// Communication Services and SQL keys.
				"Azure Identifiable Key",
				"Azure Identifiable Key",
				*regexp.MustCompile(`[A-Za-z0-9]{52}JQQJ9[9DH][A-Za-z0-9][A-L][A-Za-z0-9]{16}(?:AZAC|AZEG|AZFR|AZMR|AZMP|AZCS|msql)` +
					`[A-Za-z0-9]{4}`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
					"https://devblogs.microsoft.com/engineering-at-microsoft/common-annotated-security-keys/",
				},
			},
			{
				// 32 hex digits (AI services and API Management) assigned to an Azure
				// key name or to the subscription key header or parameter.
				"Azure API Key",
				"Azure API Key",
				*regexp.MustCompile(keyed(`azure`, `key`, `[0-9a-fA-F]{32}\b`) + `|(?i:subscription-key)[\w.-]{0,10}` + assign + `[0-9a-fA-F]{32}\b`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-cognitive-service-key",
					"https://learn.microsoft.com/en-us/azure/api-management/api-management-subscriptions",
				},
			},
			{
				// The current format: older secrets have no fixed shape.
				"Azure Entra ID Client Secret",
				"Azure Entra ID Client Secret",
				*regexp.MustCompile(`[\w~.-]{3}[78]Q~[\w~.-]{31,34}`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/PreciselyClassifiedSecurityKeys.json",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-ad-client-secret",
				},
			},
			{
				// Keys of storage and Cosmos DB accounts, Service Bus, Event Hubs, IoT,
				// SignalR, App Configuration and Cache for Redis in connection strings.
				"Azure Connection String",
				"Azure Connection String",
				*regexp.MustCompile(`(?i:AccountKey)=[A-Za-z0-9+/]{86}==|(?i:SharedAccessKey|;AccessKey)=[A-Za-z0-9+/]{43}=|` +
					`\.azconfig\.io;Id=[^;\s"'<>]{1,64};Secret=[A-Za-z0-9+/]{40,}={0,2}|` +
					`\.redis\.cache\.windows\.net(?::\d{1,5})?[,;](?:\w{1,20}=[^,;\s"'<>]{0,64}[,;]){0,8}(?i:password)=` +
					`[A-Za-z0-9+/]{43}=`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/azure/storage/common/storage-configure-connection-string",
					"https://learn.microsoft.com/en-us/azure/event-hubs/event-hubs-get-connection-string",
					"https://learn.microsoft.com/en-us/azure/azure-signalr/concept-connection-string",
					"https://learn.microsoft.com/en-us/azure/azure-app-configuration/rest-api-authentication-hmac",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-redis-cache-connection-string-password",
					"https://stackexchange.github.io/StackExchange.Redis/Configuration",
				},
			},
			{
				// Shared access signatures granting more than read, list and filter
				// (sp before sig, as the SDKs and the portal generate them).
				"Azure SAS Token",
				"Azure SAS Token",
				*regexp.MustCompile(`\bsp=[a-z]{0,20}[a-eg-km-qs-z][a-z]{0,20}(?:&|\\u0026)[^\s"'<>]{0,500}?(?:&(?:amp;)?|\\u0026)` +
					`sig=[A-Za-z0-9%/+]{40,100}`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/rest/api/storageservices/create-service-sas",
					"https://learn.microsoft.com/en-us/purview/sit-defn-azure-storage-account-shared-access-signature",
				},
			},
			{
				// Assigned to any Beamer name.
				"Beamer API Token",
				"Beamer API Token",
				*regexp.MustCompile(keyed(`beamer`, ``, `b_[\w=-]{44}`)),
				[]string{},
				"?",
				[]string{},
			},
			{
				// App passwords: ATBB and 32 characters (the length is not documented).
				"Bitbucket App Password",
				"Bitbucket App Password",
				*regexp.MustCompile(`\bATBB[A-Za-z0-9]{32}\b`),
				[]string{},
				"?",
				[]string{
					"https://support.atlassian.com/bitbucket-cloud/docs/app-passwords/",
					"https://community.atlassian.com/forums/Bitbucket-questions/Can-we-confirm-BitBucket-s-token-prefixes/qaa-p/3096469",
				},
			},
			{
				// Assigned to a secret name: commit hashes (BITBUCKET_COMMIT) look the same.
				"BitBucket Client Secret",
				"BitBucket Client Secret",
				*regexp.MustCompile(keyed(`bitbucket`, `secret`, `[A-Za-z0-9]{32,64}\b`)),
				[]string{},
				"?",
				[]string{"https://support.atlassian.com/bitbucket-cloud/docs/use-oauth-on-bitbucket-cloud/"},
			},
			{
				// 32 hex digits assigned to a private key name within 300 characters after
				// "braintree" (e.g. in the gateway configuration): the device and
				// correlation IDs look the same. The merchant IDs are info.
				"Braintree Private Key",
				"Braintree Private Key",
				*regexp.MustCompile(`(?i:braintree[\s\S]{0,300}?private_?key)[\w.-]{0,10}` + assign + `[0-9a-f]{32}\b`),
				[]string{},
				"?",
				[]string{"https://developer.paypal.com/braintree/docs/start/hello-server"},
			},
			{
				// Personal and project tokens (a base58 UUID and 40 hex digits), or the
				// older 40 hex digits assigned to a CircleCI token or key name: commit
				// hashes (CIRCLE_SHA1) look the same.
				"CircleCI Access Token",
				"CircleCI Access Token",
				*regexp.MustCompile(`\bCCIP(?:AT|RJ)_[1-9A-HJ-NP-Za-km-z]{20,24}_[0-9a-f]{40}\b|` +
					keyed(`circle`, `token|key`, `[0-9a-f]{40}\b`)),
				[]string{},
				"?",
				[]string{"https://circleci.com/changelog/new-format-for-api-access-tokens"},
			},
			{
				// User (cfut_), account (cfat_) and Wrangler OAuth (cfoat_) tokens, or the
				// older 40 characters assigned to a Cloudflare token name or to
				// CF_API_TOKEN (deprecated Wrangler name).
				"Cloudflare API Token",
				"Cloudflare API Token",
				*regexp.MustCompile(`\bcf(?:ut|at|oat)_[\w-]{40,64}|` +
					keyed(`cloudflare`, `token`, `[\w-]{40}(?:\b|[^\w-]|$)`) +
					`|(?i:\bcf_api_token)` + assign + `[\w-]{40}(?:\b|[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://developers.cloudflare.com/fundamentals/api/get-started/token-formats/",
					"https://developers.cloudflare.com/workers/wrangler/system-environment-variables/",
				},
			},
			{
				// cfk_ keys, or the older 37 to 45 hex digits assigned to a Cloudflare key
				// name or to CF_API_KEY (deprecated Wrangler name).
				"Cloudflare Global API Key",
				"Cloudflare Global API Key",
				*regexp.MustCompile(`\bcfk_[\w-]{40,64}|` +
					keyed(`cloudflare`, `key`, `[0-9a-f]{37,45}\b`) +
					`|(?i:\bcf_api_key)` + assign + `[0-9a-f]{37,45}\b`),
				[]string{},
				"?",
				[]string{
					"https://developers.cloudflare.com/fundamentals/api/get-started/token-formats/",
					"https://developers.cloudflare.com/workers/wrangler/system-environment-variables/",
				},
			},
			{
				"Cloudflare Origin CA Key",
				"Cloudflare Origin CA Key",
				*regexp.MustCompile(`\bv1\.0-[0-9a-f]{24}-[0-9a-f]{146}\b`),
				[]string{},
				"?",
				[]string{"https://developers.cloudflare.com/fundamentals/api/get-started/ca-keys/"},
			},
			{
				// UUIDs assigned to a Codecov token name (also in codecov.yml).
				"CodeCov Access Token",
				"CodeCov Access Token",
				*regexp.MustCompile(keyed(`codecov`, `token`, `[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}\b`)),
				[]string{},
				"?",
				[]string{"https://docs.codecov.com/docs/codecov-tokens"},
			},
			{
				// 40 letters and digits (the length is not documented) assigned to a Cohere
				// name (not coherent or coherence) or to CO_API_KEY, the name read by the
				// official SDK, or given to a client of the SDK.
				"Cohere API Key",
				"Cohere API Key",
				*regexp.MustCompile(keyed(`cohere(?:\b|[^n\W])|(?:\b|_)co_api_key`, ``, `[A-Za-z0-9]{40}\b`) + `|` +
					configCall(`cohere(?:\b|[^n\W])`, `api_?key|token`, `[A-Za-z0-9]{40}\b`) + `|` +
					configCall(`cohere(?:\b|[^n\W])`, ``, `[A-Za-z0-9]{40}\b`)),
				[]string{},
				"?",
				[]string{"https://docs.cohere.com/docs/rate-limits", "https://github.com/cohere-ai/cohere-python"},
			},
			{
				// Ed25519 secrets (base64) assigned to any Coinbase or CDP name: the EC keys
				// are found by the private key rule.
				"CoinBase API Key Secret",
				"CoinBase API Key Secret",
				*regexp.MustCompile(keyed(`coinbase|(?:\b|_)cdp`, ``, `[A-Za-z0-9+/]{86}==`)),
				[]string{},
				"?",
				[]string{"https://docs.cdp.coinbase.com/api-reference/authentication"},
			},
			{
				// cflt and 60 characters (with a checksum), or the older 64 characters
				// assigned to a Confluent secret name.
				"Confluent Secret Key",
				"Confluent Secret Key",
				*regexp.MustCompile(`\bcflt[A-Za-z0-9+/]{60}|` + keyed(`confluent`, `secret`, `[A-Za-z0-9+/]{64}(?:[^A-Za-z0-9+/=]|$)`)),
				[]string{},
				"?",
				[]string{
					"https://docs.confluent.io/cloud/current/security/authenticate/workload-identities/service-accounts/api-keys/overview.html",
				},
			},
			{
				// CFPAT- and 43 characters (the length is not documented): a personal access
				// token of the Content Management API.
				"Contentful Personal Access Token",
				"Contentful Personal Access Token",
				*regexp.MustCompile(`\bCFPAT-[\w-]{43}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://www.contentful.com/developers/docs/references/authentication/",
					"https://www.contentful.com/help/scim-faq/",
				},
			},
			{
				// Delivery, Preview and Management API tokens (43 characters, or 64 hex digits
				// for older ones; the format is not documented) assigned to any Contentful
				// name or to the access token of a client of the SDKs, or in the
				// access_token parameter of an API URL (also of the EU data residency):
				// read-only for delivery, but not public.
				"Contentful Access Token",
				"Contentful Access Token",
				*regexp.MustCompile(keyed(`contentful`, ``, `(?:[\w-]{43}|[0-9a-f]{64})(?:[^\w-]|$)`) + `|` +
					configCall(`contentful`, `access_?token`, `(?:[\w-]{43}|[0-9a-f]{64})(?:[^\w-]|$)`) +
					`|\b(?:api|cdn|preview|graphql)\.(?:eu\.)?contentful\.com/[^\s"'<>]{0,200}?[?&]access_token=(?:[\w-]{43}|[0-9a-f]{64})(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://www.contentful.com/developers/docs/references/authentication/",
					"https://www.contentful.com/developers/docs/infrastructure/eu-data-residency/",
				},
			},
			{
				"Databricks API Token",
				"Databricks API Token",
				*regexp.MustCompile(`\bdapi[0-9a-f]{32}(?:-\d)?\b`),
				[]string{},
				"?",
				[]string{
					"https://github.com/microsoft/security-utilities/blob/main/GeneratedRegexPatterns/HighConfidenceSecurityModels.json",
					"https://docs.databricks.com/aws/en/dev-tools/auth/pat",
				},
			},
			{
				// API keys (32 hex digits) and application keys (40) assigned to a Datadog
				// key name (also the DD-API-KEY header): commit hashes and trace IDs look
				// the same.
				"DataDog Access Token",
				"DataDog Access Token",
				*regexp.MustCompile(keyed(`datadog|(?:\b|_)dd`, `(?:api|app|application)[_.-]?key`, `[0-9a-f]{32}(?:[0-9a-f]{8})?\b`)),
				[]string{},
				"?",
				[]string{"https://docs.datadoghq.com/account_management/api-app-keys/"},
			},
			{
				// A server and a password (also in braces, see propBraced) in a connection
				// string made of properties (key=value;), in any order, as in ADO.NET, ODBC and
				// OLE DB: placeholders are not passwords, and local and documentation servers
				// are skipped (see propFalsePositive).
				"Database Connection String",
				"Database Connection String",
				*regexp.MustCompile(`\b` + propServer + `[ \t]{0,3};` + propPairs + `(?i:Password|Pwd)=\s{0,3}(?:` + propPassword + `|` + propBraced + `)` +
					`|(?i:\b(?:Password|Pwd))=\s{0,3}(?:` + propPassword + `|` + propBraced + `)[ \t]{0,3};` + propPairs + propServer),
				[]string{"example", "xxxx"},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/dotnet/framework/data/adonet/connection-string-syntax",
					"https://learn.microsoft.com/en-us/sql/connect/odbc/dsn-connection-string-attribute",
				},
			},
			{
				// sk- and 32 hex digits (the length is not documented): other providers of
				// compatible APIs use the same shape.
				"DeepSeek API Key",
				"DeepSeek API Key",
				*regexp.MustCompile(`\bsk-[0-9a-f]{32}\b`),
				[]string{},
				"?",
				[]string{"https://api-docs.deepseek.com/"},
			},
			{
				"DigitalOcean Access Token",
				"DigitalOcean Access Token",
				*regexp.MustCompile(`\bdoo_v1_[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://docs.digitalocean.com/notes/2022/api-token-format/"},
			},
			{
				"DigitalOcean Personal Access Token",
				"DigitalOcean Personal Access Token",
				*regexp.MustCompile(`\bdop_v1_[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://docs.digitalocean.com/notes/2022/api-token-format/"},
			},
			{
				"DigitalOcean Refresh Token",
				"DigitalOcean Refresh Token",
				*regexp.MustCompile(`\bdor_v1_[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://docs.digitalocean.com/notes/2022/api-token-format/"},
			},
			{
				// Tokens older than the prefixed format: 64 hex digits, assigned to a DigitalOcean
				// secret name (doctl release checksums look the same) or to a do_ token name.
				"DigitalOcean Legacy Token",
				"DigitalOcean Legacy Token",
				*regexp.MustCompile(keyed(`digital_?ocean|doctl`, secretWords+`|pat\b`, `[0-9a-f]{64}\b`) +
					`|(?i:(?:\b|_)do_(?:token|pat|api_?(?:key|token)|access_?token|auth_?token))[\w.-]{0,10}` + assign + `[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://docs.digitalocean.com/notes/2022/api-token-format/"},
			},
			{
				// The user ID (base64), a timestamp and a signature, after Bot or assigned to
				// any Discord name.
				"Discord Bot Token",
				"Discord Bot Token",
				*regexp.MustCompile(`\bBot [MNO][\w-]{23,27}\.[\w-]{6}\.[\w-]{27,40}|` + keyed(`discord`, ``, `[MNO][\w-]{23,27}\.[\w-]{6}\.[\w-]{27,40}`)),
				[]string{},
				"?",
				[]string{"https://discord.com/developers/docs/reference"},
			},
			{
				// Assigned to a secret name: avatar and icon hashes look the same.
				"Discord Client Secret",
				"Discord Client Secret",
				*regexp.MustCompile(keyed(`discord`, `secret`, `[\w-]{32}(?:\b|[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{"https://docs.discord.com/developers/topics/oauth2"},
			},
			{
				// dckr_pat_ and dckr_oat_: the length is not documented, 27 characters in the API
				// reference and 32 in some organization tokens.
				"Docker Access Token",
				"Docker Access Token",
				*regexp.MustCompile(`\bdckr_(?:pat|oat)_[\w-]{27}(?:[\w-]{5})?(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://docs.docker.com/security/access-tokens/",
					"https://docs.docker.com/reference/api/hub/latest/schemas/createOrgAccessTokenResponse.md",
				},
			},
			{
				// A UUID assigned to a Docker secret name: run and build IDs look the same.
				"Docker Legacy Access Token",
				"Docker Legacy Access Token",
				*regexp.MustCompile(keyed(`docker`, secretWords+`|pat\b`, `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)),
				[]string{},
				"?",
				[]string{"https://docs.docker.com/reference/cli/docker/login/"},
			},
			{
				// The auth field (user:password in base64) of a registry in a Docker config file,
				// also as JSON in a string. A value of 1000 characters or more has no end check:
				// regexes are limited to that.
				"Docker Registry Credentials",
				"Docker Registry Credentials",
				*regexp.MustCompile(`"auths\\*"` + jsonSpace + `:` + jsonSpace + `\{` + jsonSpace + `\\*"[\w.:/\\-]{1,200}"` + jsonSpace + `:` + jsonSpace + `\{[^}]{0,300}?` +
					`"auth\\*"` + jsonSpace + `:` + jsonSpace + `\\*"(?:[A-Za-z0-9+/]{16,999}={0,2}\\*"|[A-Za-z0-9+/]{1000})`),
				[]string{"xxxx"},
				"?",
				[]string{
					"https://docs.docker.com/reference/cli/docker/login/",
					"https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/",
				},
			},
			{
				// Short-lived access tokens (sl.), or long-lived tokens (64 characters) and app
				// secrets (15) assigned to a Dropbox token or secret name: the app keys look
				// the same.
				"DropBox API Token",
				"DropBox API Token",
				*regexp.MustCompile(`\bsl\.(?:u\.)?[\w-]{130,1000}|` +
					keyed(`dropbox`, `token|secret`, `(?:[\w-]{64}|[a-z0-9]{15})(?:\b|[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{"https://developers.dropbox.com/oauth-guide"},
			},
			{
				// sk_ and 48 hex digits (the length is not documented), also with the suffix of
				// a data residency region, or the older 32 hex digits assigned to an
				// ElevenLabs name or to the xi-api-key header.
				"ElevenLabs API Key",
				"ElevenLabs API Key",
				*regexp.MustCompile(`\bsk_[0-9a-f]{48}(?:_residency_[a-z0-9]{1,10})?(?:[^\w-]|$)|` +
					keyed(`elevenlabs|(?:\b|_)xi[_-]api[_-]key`, ``, `[0-9a-f]{32}\b`)),
				[]string{},
				"?",
				[]string{
					"https://elevenlabs.io/docs/api-reference/authentication",
					"https://www.npmjs.com/package/@elevenlabs/elevenlabs-js",
				},
			},
			{
				// Not EAAA: in base64 data, runs of A are zero bytes.
				"Facebook Access Token",
				"Facebook Access Token",
				*regexp.MustCompile(`\bEAA[B-Za-z0-9][A-Za-z0-9]{99,}`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Assigned to a Facebook secret name: client tokens, public, look the same.
				"Facebook Secret Key",
				"Facebook Secret Key",
				*regexp.MustCompile(keyed(`facebook|fb`, `secret`, `[0-9a-f]{32}\b`)),
				[]string{},
				"?",
				[]string{},
			},
			{
				// The app ID and the app secret, assigned to a token name (e.g. access_token):
				// alone, the shape is also a timestamp and a hash.
				"Facebook App Access Token",
				"Facebook App Access Token",
				*regexp.MustCompile(`(?i:(?:access|app|fb|facebook)[\w.:-]{0,30}?token)[\w.-]{0,10}` + assign +
					`\d{15,16}(?:\||%7[Cc])[0-9a-f]{32}\b`),
				[]string{},
				"?",
				[]string{"https://developers.facebook.com/docs/facebook-login/guides/access-tokens"},
			},
			{
				// Assigned to any Fastly name: the IDs have 22 characters.
				"Fastly API Token",
				"Fastly API Token",
				*regexp.MustCompile(keyed(`fastly`, ``, `[\w-]{32}(?:\b|[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{"https://docs.fastly.com/en/guides/using-api-tokens"},
			},
			{
				// Personal access tokens (figd_) and OAuth tokens (figu_, figo_, also the
				// refresh ones), or the older number-UUID tokens assigned to any Figma name.
				// The prefixes and the lengths are not documented.
				"Figma Access Token",
				"Figma Access Token",
				*regexp.MustCompile(`\bfig(?:d|p|[uo][rh]?)_[\w-]{40,54}(?:[^\w-]|$)|` + keyed(`figma`, ``, `\d{4,6}-`+uuid+`\b`)),
				[]string{},
				"?",
				[]string{"https://developers.figma.com/docs/rest-api/authentication/"},
			},
			{
				// The API key, the API secret and the cloud name (the scheme in any case).
				"Cloudinary Basic Auth",
				"Cloudinary Basic Auth",
				*regexp.MustCompile(`(?i:cloudinary)://\d{12,18}:[\w-]{20,40}@[\w-]{1,64}`),
				[]string{},
				"?",
				[]string{"https://cloudinary.com/documentation/node_integration"},
			},
			{
				// A user and a password in a URL of any scheme (user:password@host), also
				// with JSON-escaped slashes: templates (${PASSWORD}, <password>, also
				// HTML-escaped) and masks (***) are not passwords, and hosts reserved
				// for documentation and schemes with a rule of their own are skipped
				// (see exampleHost and ownScheme). The scheme has no length cap on
				// purpose: with one, the engine would scan ~1 KB around every "://".
				"Credentials in URL",
				"Credentials in URL",
				*regexp.MustCompile(`[A-Za-z][\w+.-]*:\\?/\\?/[\w.~%+-]{1,64}:[\w.~%!+,=-]{1,128}@[A-Za-z0-9](?:[A-Za-z0-9.-]{0,251}[A-Za-z0-9])?` +
					`(?::\d{1,5})?`),
				[]string{},
				"?",
				[]string{"https://www.rfc-editor.org/rfc/rfc3986#section-3.2.1", "https://www.rfc-editor.org/rfc/rfc2606"},
			},
			{
				// API key and access token secrets assigned to a Twitter secret name.
				"Twitter Secret Key",
				"Twitter Secret Key",
				*regexp.MustCompile(keyed(`twitter`, `secret`, `[A-Za-z0-9]{40,50}\b`)),
				[]string{},
				"?",
				[]string{},
			},
			{
				// The user ID and the token, assigned to any Twitter name.
				"Twitter Access Token",
				"Twitter Access Token",
				*regexp.MustCompile(keyed(`twitter`, ``, `\d{5,20}-[A-Za-z0-9]{20,50}\b`)),
				[]string{},
				"?",
				[]string{},
			},
			{
				// App-only tokens (21 A, then URL-encoded base64) after Bearer or assigned to
				// any Twitter name: runs of A are common in base64 data.
				"Twitter Bearer Token",
				"Twitter Bearer Token",
				*regexp.MustCompile(`(?i:\bbearer\s{1,5})AAAAAAAAAAAAAAAAAAAAA[A-Za-z0-9%]{30,200}|` +
					keyed(`twitter`, ``, `AAAAAAAAAAAAAAAAAAAAA[A-Za-z0-9%]{30,200}`)),
				[]string{},
				"?",
				[]string{"https://docs.x.com/fundamentals/authentication/oauth-2-0/bearer-tokens"},
			},
			{
				"Github Personal Access Token",
				"Github Personal Access Token",
				*regexp.MustCompile(`ghp_[A-Za-z0-9]{36,}`),
				[]string{},
				"?",
				[]string{
					"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github",
					"https://github.blog/engineering/platform-security/behind-githubs-new-authentication-token-formats/",
				},
			},
			{
				"Github Fine-grained Personal Access Token",
				"Github Fine-grained Personal Access Token",
				*regexp.MustCompile(`github_pat_[A-Za-z0-9]{22}_[A-Za-z0-9]{59}`),
				[]string{},
				"?",
				[]string{
					"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github",
				},
			},
			{
				"Github OAuth Access Token",
				"Github OAuth Access Token",
				*regexp.MustCompile(`gho_[A-Za-z0-9]{36,}`),
				[]string{},
				"?",
				[]string{
					"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github",
					"https://github.blog/engineering/platform-security/behind-githubs-new-authentication-token-formats/",
				},
			},
			{
				"Github App Token",
				"Github App Token",
				*regexp.MustCompile(`ghu_[A-Za-z0-9_-](?:\.?[A-Za-z0-9_-]){35,}`),
				[]string{},
				"?",
				[]string{
					"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github",
					"https://github.blog/engineering/platform-security/behind-githubs-new-authentication-token-formats/",
				},
			},
			{
				"Github Refresh Token",
				"Github Refresh Token",
				*regexp.MustCompile(`ghr_[A-Za-z0-9]{36,}`),
				[]string{},
				"?",
				[]string{
					"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github",
					"https://github.blog/engineering/platform-security/behind-githubs-new-authentication-token-formats/",
				},
			},
			{
				// Tokens and OAuth app client secrets without prefix (40 hex digits),
				// assigned to a token or secret name: commit hashes (GITHUB_SHA) look the same.
				"Github Legacy Secret",
				"Github Legacy Secret",
				*regexp.MustCompile(keyed(`github|gh_`, `token|secret|pat\b|api_?key|access_?key`, `[0-9a-f]{40}\b`)),
				[]string{},
				"?",
				[]string{"https://github.blog/engineering/platform-security/behind-githubs-new-authentication-token-formats/"},
			},
			{
				"Github App Installation Token",
				"Github App Installation Token (https://github.blog/changelog/2026-05-15-github-app-installation-tokens-per-request-override-header/)",
				*regexp.MustCompile(`ghs_[A-Za-z0-9_-](?:\.?[A-Za-z0-9_-]){35,}`),
				[]string{},
				"?",
				[]string{
					"https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/about-authentication-to-github",
					"https://github.blog/changelog/2026-05-15-github-app-installation-tokens-per-request-override-header/",
				},
			},
			{
				"GitLab CI/CD Job Token",
				"GitLab CI/CD Job Token",
				*regexp.MustCompile(`glcbt-[0-9a-zA-Z]{1,5}_[0-9a-zA-Z_-]{20}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab Deploy Token",
				"GitLab Deploy Token",
				*regexp.MustCompile(`gldt-[\w-]{20,300}(?:\.[0-9a-z]{2}\.[0-9a-z]{9})?`),
				[]string{},
				"?",
				[]string{
					"https://docs.gitlab.com/security/tokens/",
					"https://handbook.gitlab.com/handbook/engineering/architecture/design-documents/cells/routable_tokens/",
				},
			},
			{
				"GitLab feature flag client token",
				"GitLab feature flag client token",
				*regexp.MustCompile(`glffct-[0-9a-zA-Z_\-]{20}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab Feed Token",
				"GitLab Feed Token",
				*regexp.MustCompile(`glft-[0-9a-zA-Z_\-]{20}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab Incoming Mail Token",
				"GitLab Incoming Mail Token",
				*regexp.MustCompile(`glimt-[0-9a-zA-Z_\-]{25}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab Kubernetes Agent Token",
				"GitLab Kubernetes Agent Token",
				*regexp.MustCompile(`glagent-[0-9a-zA-Z_\-]{50}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab OIDC Application Secret",
				"GitLab OIDC Application Secret",
				*regexp.MustCompile(`gloas-[0-9a-zA-Z_\-]{64}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				// Also routable: base64url payload, then .version.length+checksum.
				"GitLab Personal Access Token",
				"GitLab Personal Access Token",
				*regexp.MustCompile(`glpat-[\w-]{20,300}(?:\.[0-9a-z]{2}\.[0-9a-z]{9})?`),
				[]string{},
				"?",
				[]string{
					"https://docs.gitlab.com/security/tokens/",
					"https://handbook.gitlab.com/handbook/engineering/architecture/design-documents/cells/routable_tokens/",
				},
			},
			{
				// Assigned to a GitLab token name: 20 characters without prefix, like many
				// user and project names.
				"GitLab Legacy Personal Access Token",
				"GitLab Legacy Personal Access Token",
				*regexp.MustCompile(keyed(`gitlab`, `token|pat\b|secret`, `[A-Za-z0-9_-]{20}(?:\b|[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{},
			},
			{
				"GitLab Pipeline Trigger Token",
				"GitLab Pipeline Trigger Token",
				*regexp.MustCompile(`glptt-[0-9a-f]{40}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab Runner Authentication Token",
				"GitLab Runner Authentication Token",
				*regexp.MustCompile(`glrtr?-[\w-]{20,300}(?:\.[0-9a-z]{2}\.[0-9a-z]{9})?`),
				[]string{},
				"?",
				[]string{
					"https://docs.gitlab.com/security/tokens/",
					"https://handbook.gitlab.com/handbook/engineering/architecture/design-documents/cells/routable_tokens/",
				},
			},
			{
				"GitLab Runner Registration Token",
				"GitLab Runner Registration Token",
				*regexp.MustCompile(`\bGR1348941[\w-]{20}`),
				[]string{},
				"?",
				[]string{"https://gitlab.com/gitlab-org/gitlab/-/blob/master/app/models/concerns/runners_token_prefixable.rb"},
			},
			{
				"GitLab Session Cookie",
				"GitLab Session Cookie",
				*regexp.MustCompile(`_gitlab_session=[\w-]{32,}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				"GitLab SCIM Token",
				"GitLab SCIM Token",
				*regexp.MustCompile(`glsoat-[0-9a-zA-Z_\-]{20}`),
				[]string{},
				"?",
				[]string{"https://docs.gitlab.com/security/tokens/"},
			},
			{
				// The older API keys: the key (32 characters), the name and the organization ID,
				// as JSON in base64 ({"k":"...","n":"...","id":1}).
				"Grafana API Key",
				"Grafana API Key",
				*regexp.MustCompile(`\beyJrIjoi[A-Za-z0-9+/]{43}iLCJuIjoi[A-Za-z0-9+/]{2,300}={0,2}`),
				[]string{},
				"?",
				[]string{
					"https://grafana.com/docs/grafana/latest/administration/service-accounts/migrate-api-keys/",
					"https://github.com/grafana/grafana/blob/v9.5.0/pkg/components/apikeygen/apikeygen.go",
				},
			},
			{
				"Grafana Service Account Token",
				"Grafana Service Account Token",
				*regexp.MustCompile(`\bglsa_[A-Za-z0-9]{32}_[A-Fa-f0-9]{8}\b`),
				[]string{},
				"?",
				[]string{"https://grafana.com/docs/grafana/latest/administration/service-accounts/"},
			},
			{
				"Grafana Cloud Access Policy Token",
				"Grafana Cloud Access Policy Token",
				*regexp.MustCompile(`\bglc_[A-Za-z0-9/+]{32,400}={0,2}`),
				[]string{},
				"?",
				[]string{
					"https://grafana.com/docs/learning-paths/connect-agent-to-agent-observability/create-access-policy-and-token/",
				},
			},
			{
				// gsk_ and 52 letters and digits (the format is not documented).
				"Groq API Key",
				"Groq API Key",
				*regexp.MustCompile(`\bgsk_[A-Za-z0-9]{52}\b`),
				[]string{},
				"?",
				[]string{"https://console.groq.com/docs/quickstart"},
			},
			{
				// HCP Terraform and Terraform Enterprise tokens: 14 characters, .atlasv1. and
				// 60 to 70 characters (the lengths are not documented).
				"HashiCorp Terraform API Token",
				"HashiCorp Terraform API Token",
				*regexp.MustCompile(`\b[A-Za-z0-9]{14}\.atlasv1\.[A-Za-z0-9]{60,70}\b`),
				[]string{},
				"?",
				[]string{"https://developer.hashicorp.com/terraform/cli/config/config-file"},
			},
			{
				// Service (hvs.), batch (hvb.) and recovery (hvr.) tokens, or the ones of
				// Vault 1.9 and earlier (s., b. and r.) assigned to any Vault name, e.g.
				// VAULT_TOKEN or the X-Vault-Token header.
				"HashiCorp Vault Token",
				"HashiCorp Vault Token",
				*regexp.MustCompile(`\bhv[sbr]\.[\w-]{24,300}|` +
					keyed(`vault`, ``, `(?:[sr]\.[A-Za-z0-9]{24}|b\.[\w-]{100,300})(?:["'\x60\s,;<&]|\\[nrt]|$)`)),
				[]string{},
				"?",
				[]string{"https://developer.hashicorp.com/vault/docs/concepts/tokens"},
			},
			{
				// The unseal and recovery keys printed when a Vault is initialized, as text
				// (Unseal Key 1: ...) or JSON, in base64 or hex.
				"HashiCorp Vault Unseal Key",
				"HashiCorp Vault Unseal Key",
				*regexp.MustCompile(`\b(?:Unseal|Recovery) Key \d{1,2}: [A-Za-z0-9+/]{43,44}=?(?:[^A-Za-z0-9+/=]|$)|` +
					`"(?:(?:unseal|recovery)_keys_(?:b64|hex)|keys_base64)"\s{0,5}:\s{0,5}\[\s{0,20}"(?:[A-Za-z0-9+/]{43,44}=?|[0-9a-f]{64,66})"`),
				[]string{},
				"?",
				[]string{
					"https://developer.hashicorp.com/vault/docs/commands/operator/init",
					"https://developer.hashicorp.com/vault/api-docs/system/init",
				},
			},
			{
				// The SecretID of an AppRole (a UUID), assigned to a Vault or AppRole
				// secret ID name, or after its RoleID (e.g. in a login): alone, secret_id
				// names any ID. The RoleIDs are info.
				"HashiCorp Vault AppRole Secret ID",
				"HashiCorp Vault AppRole Secret ID",
				*regexp.MustCompile(keyed(`vault|approle`, `secret[_-]?id`, uuid+`\b`) +
					`|(?i:\brole_?id)` + assign + uuid + `\b[\s\S]{0,60}?(?i:\bsecret_?id)` + assign + uuid + `\b`),
				[]string{},
				"?",
				[]string{"https://developer.hashicorp.com/vault/docs/auth/approle"},
			},
			{
				// Private app access tokens: pat-, the data center and a UUID.
				"HubSpot Private App Access Token",
				"HubSpot Private App Access Token",
				*regexp.MustCompile(`\bpat-[a-z]{2}\d-` + uuid + `\b`),
				[]string{},
				"?",
				[]string{"https://developers.hubspot.com/docs/guides/apps/private-apps/overview"},
			},
			{
				// Developer API keys and app client secrets (UUIDs) assigned to a HubSpot
				// key or secret name, or to hapikey (also in a URL): the form, portal and app
				// IDs look the same. The account API keys stopped working in 2022.
				"HubSpot API Key",
				"HubSpot API Key",
				*regexp.MustCompile(keyed(`hubspot|hubapi`, `key|secret`, uuid+`\b`) + `|(?i:\bhapi_?key)[\w.-]{0,10}` + assign + uuid + `\b`),
				[]string{},
				"?",
				[]string{"https://developers.hubspot.com/docs/guides/apps/authentication/intro-to-auth"},
			},
			{
				// The length is not documented: 34 letters and digits.
				"Hugging Face Access Token",
				"Hugging Face Access Token",
				*regexp.MustCompile(`\bhf_[A-Za-z0-9]{34}\b`),
				[]string{},
				"?",
				[]string{"https://huggingface.co/docs/hub/security-tokens"},
			},
			{
				// The deprecated organization API tokens: the length is not documented.
				"Hugging Face Organization API Token",
				"Hugging Face Organization API Token",
				*regexp.MustCompile(`\bapi_org_[A-Za-z0-9]{34}\b`),
				[]string{},
				"?",
				[]string{"https://huggingface.co/docs/hub/security-tokens"},
			},
			{
				// 44 characters assigned to an IBM Cloud (or Bluemix) name or to IC_API_KEY,
				// after --apikey in a login command, given to an IAM authenticator of the
				// SDKs, or in the apikey field of service credentials (with their iam_
				// fields). A key ID (ApiKey- and a UUID) has 43.
				"IBM Cloud API Key",
				"IBM Cloud API Key",
				*regexp.MustCompile(keyed(`(?:\b|_)(?:ibm|bluemix|ic_api_key)`, ``, `[\w-]{44}(?:[^\w-]|$)`) +
					`|(?i:\b(?:ibmcloud|bluemix)\s{1,5}login\b[^\n]{0,100}?\s--?apikey)[ =]["']?[\w-]{44}(?:[^\w-]|$)` +
					`|(?i:\bIamAuthenticator\b)[^"'\x60;]{0,40}["'\x60][\w-]{44}["'\x60]` +
					`|\\?"apikey\\?"` + jsonSpace + `:` + jsonSpace + `\\?"[\w-]{44}\\?"[^\]]{0,600}?\\?"iam_(?:apikey_description|apikey_name|role_crn|serviceid_crn)\\?"`),
				[]string{},
				"?",
				[]string{
					"https://cloud.ibm.com/docs/account?topic=account-userapikey",
					"https://cloud.ibm.com/docs/cli?topic=cli-ibmcloud_env_var",
					"https://cloud.ibm.com/docs/cloud-object-storage?topic=cloud-object-storage-service-credentials",
					"https://github.com/IBM/ibm-cloud-sdk-common",
					"https://registry.terraform.io/providers/IBM-Cloud/ibm/latest/docs",
				},
			},
			{
				// API keys (32 hex digits, the format is not documented) assigned to any
				// Infura name, as INFURA_API_KEY, given to a call on one, as
				// ethers.InfuraProvider(network, key), or in the path of an endpoint
				// (https://<network>.infura.io/v3/<key>, also for WebSocket and the Gas API):
				// a key works as a password. The API key secrets look the same and are found
				// by their names, as INFURA_API_KEY_SECRET.
				"Infura API Key",
				"Infura API Key",
				*regexp.MustCompile(keyed(`infura`, ``, `[0-9a-f]{32}\b`) + `|(?i:infura[\w.]{0,30}\s{0,5}\([^()]{0,100}?)["'\x60][0-9a-f]{32}\b` +
					`|\b[a-z0-9-]{1,63}\.infura\.io/(?:ws/)?v3/[0-9a-f]{32}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.infura.io/dashboard/how-to/secure-an-api/api-key",
					"https://docs.infura.io/dashboard/how-to/secure-an-api/api-key-secret",
				},
			},
			{
				// IGQVJ and the rest of a user access token of the Instagram API.
				"Instagram Access Token",
				"Instagram Access Token",
				*regexp.MustCompile(`\bIGQVJ[\w-]{40,1000}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{"https://developers.facebook.com/docs/instagram-platform/reference/access_token"},
			},
			{
				// Access tokens: tok:, an ID and a version, in base64 (60 characters). The
				// format is not documented.
				"Intercom Access Token",
				"Intercom Access Token",
				*regexp.MustCompile(`\bdG9rO[A-Za-z0-9+/]{54}=`),
				[]string{},
				"?",
				[]string{"https://developers.intercom.com/docs/build-an-integration/learn-more/authentication"},
			},
			{
				// A password in the properties of a JDBC URL (?user=...&password=... or
				// ;user=...;password=...), or user/password@ in an Oracle one: placeholders are
				// not passwords, and local and documentation hosts are skipped (see
				// jdbcFalsePositive). The URLs with user:password@host are found by the other
				// connection string and URL rules.
				"JDBC Connection String",
				"JDBC Connection String",
				*regexp.MustCompile(`(?i:jdbc:[a-z][a-z0-9+.-]{1,30}:)[^\s"'\x60<>]{1,300}?(?i:[?&;:](?:password|pwd))=` + propPassword +
					`|(?i:jdbc:oracle:[a-z0-9]{1,10}:)` + connUser + `/` + connPassword + `@(?:(?:\\?/\\?/)?[A-Za-z0-9][A-Za-z0-9.-]{0,252}|\()`),
				[]string{"example", "xxxx", "%2a%2a"},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/sql/connect/jdbc/building-the-connection-url",
					"https://jdbc.postgresql.org/documentation/use/",
					"https://docs.oracle.com/en/database/oracle/oracle-database/19/jjdbc/data-sources-and-URLs.html",
				},
			},
			{
				// 11 and 32 hex digits (or the older 32 hex digits) assigned to a Jenkins key,
				// token, secret or password name: the crumbs look the same.
				"Jenkins API Token",
				"Jenkins API Token",
				*regexp.MustCompile(keyed(`jenkins`, secretWords, `(?:11)?[0-9a-f]{32}\b`)),
				[]string{},
				"?",
				[]string{
					"https://www.jenkins.io/blog/2018/07/02/new-api-token-system/",
					"https://github.com/jenkinsci/jenkins/blob/master/core/src/main/java/jenkins/security/apitoken/ApiTokenStore.java",
					"https://www.jenkins.io/doc/book/security/csrf-protection/",
				},
			},
			{
				// The password of the first admin user, printed in the log by the setup wizard.
				"Jenkins Setup Admin Password",
				"Jenkins Setup Admin Password",
				*regexp.MustCompile(`Please use the following password to proceed to installation:(?:\s|\\[rn]){1,10}[0-9a-f]{32}\b`),
				[]string{},
				"?",
				[]string{
					"https://www.jenkins.io/doc/book/installing/linux/",
					"https://github.com/jenkinsci/jenkins/blob/master/core/src/main/java/jenkins/install/SetupWizard.java",
				},
			},
			{
				// AKCp and 69 letters and digits.
				"JFrog API Key",
				"JFrog API Key",
				*regexp.MustCompile(`\bAKCp[A-Za-z0-9]{69}\b`),
				[]string{},
				"?",
				[]string{"https://docs.jfrog.com/user-management/docs/identity-tokens"},
			},
			{
				// cmVmd (the start of reftkn in base64) and 59 letters and digits: identity
				// and access tokens as references (the full tokens are JWTs).
				"JFrog Reference Token",
				"JFrog Reference Token",
				*regexp.MustCompile(`\bcmVmd[A-Za-z0-9]{59}\b`),
				[]string{},
				"?",
				[]string{"https://docs.jfrog.com/user-management/docs/identity-tokens"},
			},
			{
				// A signed JWT: the header, the payload (base64url-encoded JSON objects, eyJ
				// is {", ewo is {\n) and the signature.
				"JSON Web Token",
				"JSON Web Token",
				*regexp.MustCompile(`(?:eyJ|ewo)[\w-]{10,}\.(?:eyJ|ewo)[\w-]{10,}\.[\w-]{20,}`),
				[]string{},
				"?",
				[]string{
					"https://www.rfc-editor.org/rfc/rfc7519",
					"https://www.rfc-editor.org/rfc/rfc7515#section-7.1",
				},
			},
			{
				// The base64 of a JWT: of a header starting with "alg", "typ" or "kid", then
				// of ".eyJ" (in one of its three alignments) and of the rest.
				"JSON Web Token (base64)",
				"JSON Web Token (base64)",
				*regexp.MustCompile(`\bZXlK(?:aGJHY2lPa|MGVYQWlPa|cmFXUWlPa)[A-Za-z0-9+/]{10,}?(?:LmV5S|[CSiy]5leU[o-r]|uZXlK)` +
					`[A-Za-z0-9+/]{40,1000}`),
				[]string{},
				"?",
				[]string{"https://www.rfc-editor.org/rfc/rfc7519"},
			},
			{
				// The signing secret assigned to a JWT secret name: quoted, or alone up to the
				// end of the line, not a reference to a variable, a file or a template.
				"JSON Web Token Secret",
				"JSON Web Token Secret",
				*regexp.MustCompile(keyed(`jwt`, `secret`, `(?:[^\s"'\x60$<>{}()\[\]\\/.][^\s"'\x60$<>{}()\[\]\\]{7,127}["'\x60]|`+
					`[\w!#%&*+:=?@^|~-][\w!#%&*+/:=?@^|~-]{7,127}[ \t]*(?:\r?\n|$))`)),
				[]string{},
				"?",
				[]string{"https://www.rfc-editor.org/rfc/rfc7518#section-3.2"},
			},
			{
				// A token ID (6 characters), a dot and a secret (16), after a token or
				// bootstrap name (also kubeadm join --token) or Bearer, or the token-id and
				// token-secret of a bootstrap token Secret.
				"Kubernetes Bootstrap Token",
				"Kubernetes Bootstrap Token",
				*regexp.MustCompile(`(?:(?i:token|bootstrap)[\w.-]{0,20}(?:["']?\s{0,5}[:=]{1,2}\s{0,5}|\s{1,5})["']?|(?i:bearer)\s{1,5})` +
					`[a-z0-9]{6}\.[a-z0-9]{16}(?:[^\w.-]|$)` +
					`|\btoken-id:[ \t]{0,5}["']?[a-z0-9]{6}["']?[ \t]{0,5}(?:\r?\n|\\[rn]){1,4}[ \t]{0,20}token-secret:[ \t]{0,5}["']?[a-z0-9]{16}\b`),
				[]string{},
				"?",
				[]string{"https://kubernetes.io/docs/reference/access-authn-authz/bootstrap-tokens/"},
			},
			{
				// A Secret manifest whose data has a value in base64 (see kubeData): in YAML,
				// kind: Secret (one space, as kubectl and the documentation write it) and then
				// data (within 12 lines, not across another document or a list item), or data
				// and then kind: Secret, as kubectl sorts the fields; in JSON (also in a
				// string), the same in either order. Base64 is not encryption: the values are
				// the secrets. Templates and stringData are not matched.
				"Kubernetes Secret",
				"Kubernetes Secret",
				*regexp.MustCompile(`\bkind: ["']?Secret["']?[ \t]{0,5}\r?\n(?:[ \t]{0,20}(?:[^-\s][^\n]{0,59})?\n){0,12}?` +
					`[ \t]{0,20}data:[ \t]{0,5}\r?\n` + kubeData + `["']?[ \t]{0,5}(?:\r?\n|$)` +
					`|\bdata:[ \t]{0,5}\r?\n` + kubeData + `["']?[ \t]{0,5}\r?\n(?:[ \t](?:[^\n]|\n[ \t]){0,1000}?\n)?` +
					`(?:[ \t]{0,20}immutable:[ \t]{0,5}(?:true|false)[ \t]{0,5}\r?\n)?[ \t]{0,20}kind: ["']?Secret["']?[ \t]{0,5}(?:\r?\n|$)` +
					`|\\?"kind\\?"` + jsonSpace + `:` + jsonSpace + `\\?"Secret\\?"(?:[^"\[\]]|"(?:[^k\[\]]|k[^i])){0,600}?` + kubeJSONData + `\\?"` +
					`|` + kubeJSONData + `\\?"[^{}\[\]]{0,1000}?\}` + jsonSpace + `,` + jsonSpace +
					`(?:\\?"immutable\\?"` + jsonSpace + `:` + jsonSpace + `(?:true|false)` + jsonSpace + `,` + jsonSpace + `)?` +
					`\\?"kind\\?"` + jsonSpace + `:` + jsonSpace + `\\?"Secret\\?"`),
				[]string{},
				"?",
				[]string{
					"https://kubernetes.io/docs/concepts/configuration/secret/",
					"https://kubernetes.io/docs/tasks/configmap-secret/managing-secret-using-config-file/",
				},
			},
			{
				// Assigned to a LinkedIn secret name: 16 characters, or the newer WPL_AP format
				// (with /, + and =, also URL-encoded).
				"LinkedIn Secret Key",
				"LinkedIn Secret Key",
				*regexp.MustCompile(keyed(`linkedin`, `secret`, `(?:WPL_AP\d\.[\w.+/%=-]{16,80}|[A-Za-z0-9]{16}\b)`)),
				[]string{},
				"?",
				[]string{"https://learn.microsoft.com/en-us/linkedin/shared/authentication/client-credentials-flow"},
			},
			{
				// OAuth access tokens (AQ...) assigned to a LinkedIn token name, and the li_at
				// session cookie.
				"LinkedIn Access Token",
				"LinkedIn Access Token",
				*regexp.MustCompile(keyed(`linkedin`, `token`, `AQ[\w-]{20,1000}`) + `|\bli_at=AQ[\w-]{50,1000}`),
				[]string{},
				"?",
				[]string{"https://learn.microsoft.com/en-us/linkedin/shared/authentication/authorization-code-flow"},
			},
			{
				"Slack",
				"Slack",
				*regexp.MustCompile(`(?:xoxe\.(?:xox[bp]|xapp)|xox[abeprs]|xapp|xwfp)-[0-9]+(?:-[0-9a-zA-Z]+)*-[0-9a-zA-Z]{6,}`),
				[]string{},
				"?",
				[]string{"https://docs.slack.dev/authentication/tokens"},
			},
			{
				// The key material is required, as the header alone is in every library
				// parsing keys. Header lines (Proc-Type, DEK-Info, Version...) can precede
				// it and newlines can be escaped (\n), as in JSON or code.
				"Asymmetric Private Key",
				"Asymmetric Private Key",
				*regexp.MustCompile(`-----BEGIN (?:[A-Z0-9]{1,12} ){0,3}PRIVATE KEY(?: BLOCK)?-----(?:\s|\\[nr]){0,32}` +
					`(?:[A-Za-z-]{1,20}: [^\n\\]{1,100}(?:\s|\\[nr]){1,32}){0,4}[A-Za-z0-9+/]{40}`),
				[]string{},
				"?",
				[]string{
					"https://www.rfc-editor.org/rfc/rfc7468",
					"https://www.rfc-editor.org/rfc/rfc1421",
					"https://github.com/openssh/openssh-portable/blob/master/PROTOCOL.key",
					"https://www.rfc-editor.org/rfc/rfc9580",
				},
			},
			{
				// A base64-encoded PEM private key (also after other blocks, e.g. a
				// certificate): "-----BEGIN", then "PRIVATE KEY", each at any of its 3
				// alignments, then key material.
				"Asymmetric Private Key (base64)",
				"Asymmetric Private Key (base64)",
				*regexp.MustCompile(`(?:LS0tLS1CRUdJTi|0tLS0tQkVHSU4g|tLS0tLUJFR0lOI)[A-Za-z0-9+/]{0,24}?` +
					`(?:UFJJVkFURSBL|BSSVZBVEUgS0VZ|QUklWQVRFIEtF)[A-Za-z0-9+/]{60}`),
				[]string{},
				"?",
				[]string{
					"https://www.rfc-editor.org/rfc/rfc7468",
					"https://www.rfc-editor.org/rfc/rfc4648#section-4",
				},
			},
			{
				"Google API key",
				"Google API key",
				*regexp.MustCompile(`\bAIza[\w-]{35}`),
				[]string{},
				"?",
				[]string{"https://docs.cloud.google.com/docs/authentication/api-keys"},
			},
			{
				// The base64 of "AIzaSy" and the rest of the key.
				"Google API key (base64)",
				"Google API key (base64)",
				*regexp.MustCompile(`QUl6YVN5[A-Za-z0-9+/]{44}`),
				[]string{},
				"?",
				[]string{"https://docs.cloud.google.com/docs/authentication/api-keys"},
			},
			{
				"Google (GCP) Service Account",
				"Google (GCP) Service Account",
				*regexp.MustCompile(`"type"\s{0,5}:\s{0,5}"service_account"`),
				[]string{},
				"?",
				[]string{
					"https://docs.cloud.google.com/iam/docs/keys-create-delete",
					"https://google.aip.dev/auth/4112",
				},
			},
			{
				// The credentials file of a user (client secret and refresh token).
				"Google Application Default Credentials",
				"Google Application Default Credentials",
				*regexp.MustCompile(`"type"\s{0,5}:\s{0,5}"authorized_user"`),
				[]string{},
				"?",
				[]string{"https://google.aip.dev/auth/4113"},
			},
			{
				// GOCSPX- secrets, or older ones assigned to client_secret after a client ID.
				"Google OAuth Client Secret",
				"Google OAuth Client Secret",
				*regexp.MustCompile(`\bGOCSPX-[A-Za-z0-9_-]{28,100}|\d{1,30}-[0-9a-z]{32}\.apps\.googleusercontent\.com[\s\S]{0,200}?` +
					`(?i:client_?secret)` + assign + `[A-Za-z0-9_-]{24}(?:\b|[^\w-]|$)`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Legacy server key: AAAA, 7 characters, then the same body of device tokens.
				"Firebase Cloud Messaging Server Key",
				"Firebase Cloud Messaging Server Key",
				*regexp.MustCompile(`\bAAAA[\w-]{7}:APA91b[\w-]{100,}`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// UUIDs or 40 hex digits assigned to a Heroku key or token name (the app and
				// dyno IDs are UUIDs too), or the password of a Heroku machine in .netrc.
				"Heroku API key",
				"Heroku API key",
				*regexp.MustCompile(keyed(`heroku`, `key|token|secret`, `(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[0-9a-f]{40})\b`) +
					`|machine\s{1,5}(?:api|git)\.heroku\.com\s{1,10}login\s{1,5}\S{1,100}\s{1,10}password\s{1,5}` +
					`(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[0-9a-f]{40})\b`),
				[]string{},
				"?",
				[]string{"https://devcenter.heroku.com/articles/authentication"},
			},
			{
				"Heroku API Key v2",
				"Heroku API Key v2",
				*regexp.MustCompile(`\bHRKU-(?:AA[\w-]{58}|[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})`),
				[]string{},
				"?",
				[]string{"https://devcenter.heroku.com/changelog-items/2800"},
			},
			{
				// 32 hex digits and the data center.
				"MailChimp API key",
				"MailChimp API key",
				*regexp.MustCompile(`\b[0-9a-f]{32}-us\d{1,2}\b`),
				[]string{},
				"?",
				[]string{"https://mailchimp.com/help/about-api-keys/"},
			},
			{
				// Older keys (key- and 32 characters), and the current ones (32-8-8 hex digits)
				// up to 200 characters after mailgun (e.g. in its config), assigned to an
				// MG_API name, or after api: (the basic auth user).
				"Mailgun API key",
				"Mailgun API key",
				*regexp.MustCompile(`\bkey-[0-9a-z]{32}\b|(?:(?i:mailgun)[\s\S]{0,200}?|(?i:(?:\b|_)mg_?api[\w.-]{0,20}?(?:key|secret))` +
					`[\w.-]{0,10}` + assign + `|\bapi:)\b[0-9a-f]{32}-[0-9a-f]{8}-[0-9a-f]{8}\b`),
				[]string{},
				"?",
				[]string{"https://documentation.mailgun.com/docs/mailgun/api-reference/authentication/"},
			},
			{
				// 32 characters assigned to a Mistral key, token or secret name: the
				// checksums of the model files look the same.
				"Mistral API Key",
				"Mistral API Key",
				*regexp.MustCompile(keyed(`mistral`, `key|token|secret`, `[A-Za-z0-9]{32}\b`)),
				[]string{},
				"?",
				[]string{"https://docs.mistral.ai/getting-started/quickstart"},
			},
			{
				// The API secret of a project (deprecated) or the secret of a service account:
				// 32 letters and digits (the format is not documented) assigned to a Mixpanel
				// secret or password name or given to an SDK as its secret (the project token,
				// public, looks the same), or after the username of a service account
				// (<name>.<6 hex digits>.mp-service-account), after a colon (Basic auth) or
				// as the next quoted argument.
				"Mixpanel Secret",
				"Mixpanel Secret",
				*regexp.MustCompile(keyed(`mixpanel`, `secret|pass|pwd`, `[A-Za-z0-9]{32}\b`) + `|` +
					configCall(`mixpanel`, `(?:api_)?secret`, `[A-Za-z0-9]{32}\b`) +
					`|\b[\w-]{1,100}\.[0-9a-f]{6}\.mp-service-account(?::|["']\s{0,20},\s{0,20}(?:secret\s{0,5}=\s{0,5})?["'])[A-Za-z0-9]{32}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.mixpanel.com/reference/project-secret",
					"https://docs.mixpanel.com/reference/service-accounts",
					"https://docs.mixpanel.com/reference/project-token",
				},
			},
			{
				// npm_ and 36 characters, the last 6 a checksum: granular and session tokens
				// (the older UUID tokens are all revoked).
				"npm Access Token",
				"npm Access Token",
				*regexp.MustCompile(`\bnpm_[A-Za-z0-9]{36}\b`),
				[]string{},
				"?",
				[]string{
					"https://github.blog/changelog/2021-09-23-npm-has-a-new-access-token-format/",
					"https://docs.npmjs.com/about-access-tokens",
				},
			},
			{
				// The secret of an Atlas service account: the length is not documented.
				"MongoDB Atlas Service Account Secret",
				"MongoDB Atlas Service Account Secret",
				*regexp.MustCompile(`\bmdb_sa_sk_[\w-]{40}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://docs.aws.amazon.com/secretsmanager/latest/userguide/mes-partner-MongoDBAtlasServiceAccount.html",
					"https://www.mongodb.com/docs/cloud-manager/reference/api/service-accounts/org/create-service-account-secret",
				},
			},
			{
				// The private key of an Atlas API key (a UUID), assigned to an Atlas secret name,
				// also in camelCase, or in the block of the Terraform provider: encryption key
				// IDs look the same.
				"MongoDB Atlas API Private Key",
				"MongoDB Atlas API Private Key",
				*regexp.MustCompile(keyed(mongoAtlas, `priv|secret|pass|api[_-]?key`, `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`) +
					`|(?i:atlas(?:api)?(?:priv|secret))[\w.-]{0,10}` + assign + `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b` +
					`|mongodbatlas"\s{0,5}\{[^}]{0,200}?private_key\s{0,5}=\s{0,5}"[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`),
				[]string{},
				"?",
				[]string{
					"https://www.mongodb.com/docs/atlas/cli/current/atlas-cli-env-variables/",
					"https://developer.hashicorp.com/vault/docs/secrets/mongodbatlas",
				},
			},
			{
				// mongodb:// and mongodb+srv:// (or another suffix) with a user, a password
				// and a host (see connFalsePositive).
				"MongoDB Connection String",
				"MongoDB Connection String",
				*regexp.MustCompile(connectionString(`mongodb(?:\+\w{1,20})?`)),
				[]string{"example", "xxxx", "%2a%2a"},
				"?",
				[]string{"https://www.mongodb.com/docs/manual/reference/connection-string-formats/"},
			},
			{
				// With a user, a password and a host (see connFalsePositive): mysql://, mysqlx://,
				// mysqlx+srv://, mysql2:// and the mysql+driver:// of SQLAlchemy.
				"MySQL Connection String",
				"MySQL Connection String",
				*regexp.MustCompile(connectionString(`mysql[x2]?(?:\+\w{1,20})?`)),
				[]string{"example", "xxxx", "%2a%2a"},
				"?",
				[]string{
					"https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html",
					"https://docs.sqlalchemy.org/en/20/core/engines.html",
				},
			},
			{
				// Personal access (nfp_), CLI (nfc_), OAuth (nfo_), app (nfu_) and build (nfb_)
				// tokens, 40 characters, or the older ones (40 to 64 characters, not documented)
				// assigned to any Netlify name.
				"Netlify Access Token",
				"Netlify Access Token",
				*regexp.MustCompile(`\bnf[pcoub]_[\w-]{36}(?:[^\w-]|$)|` + keyed(`netlify`, ``, `[\w-]{40,64}(?:[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{
					"https://answers.netlify.com/t/change-to-the-netlify-authentication-token-format/106146",
					"https://docs.netlify.com/api-and-cli-guides/api-guides/get-started-with-api/",
				},
			},
			{
				// User keys (NRAK-), also the older admin keys (NRAA-): the length is not
				// documented.
				"New Relic User API Key",
				"New Relic User API Key",
				*regexp.MustCompile(`\bNRA[AK]-[A-Za-z0-9]{27}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/",
					"https://docs.aws.amazon.com/devopsagent/latest/APIReference/API_NewRelicApiKeyConfig.html",
				},
			},
			{
				// The older REST API keys: the length is not documented.
				"New Relic REST API Key",
				"New Relic REST API Key",
				*regexp.MustCompile(`\bNRRA-[0-9a-f]{42}\b`),
				[]string{},
				"?",
				[]string{"https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/"},
			},
			{
				// The older keys for sending data: the length is not documented.
				"New Relic Insights Insert Key",
				"New Relic Insights Insert Key",
				*regexp.MustCompile(`\bNRII-[\w-]{32}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{"https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/"},
			},
			{
				// The older keys for querying data: the length is not documented.
				"New Relic Insights Query Key",
				"New Relic Insights Query Key",
				*regexp.MustCompile(`\bNRIQ-[\w-]{32}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{"https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/"},
			},
			{
				// 40 characters ending in NRAL (also of the EU data center), or the older
				// 40 hex digits assigned to a New Relic license name or to license_key within
				// 300 characters after "New Relic" (as in the configuration of the agents).
				"New Relic License Key",
				"New Relic License Key",
				*regexp.MustCompile(`\b(?:[0-9a-f]{36}|eu01xx[0-9a-f]{30})NRAL\b|` + keyed(`new[_-]?relic`, `license`, `[0-9a-f]{40}\b`) +
					`|(?i:new[ _-]?relic[\s\S]{0,300}?\blicense_key)` + assign + `[0-9a-f]{40}\b`),
				[]string{},
				"?",
				[]string{"https://docs.newrelic.com/docs/apis/intro-apis/new-relic-api-keys/"},
			},
			{
				// API (px-api-) and deployment (px-dep-) keys of Pixie.
				"New Relic Pixie Key",
				"New Relic Pixie Key",
				*regexp.MustCompile(`\bpx-(?:api|dep)-` + uuid + `\b`),
				[]string{},
				"?",
				[]string{"https://docs.px.dev/reference/admin/api-keys/", "https://docs.px.dev/reference/admin/deploy-keys/"},
			},
			{
				// ntn_ (11 digits and 35 characters), or secret_ and 43 characters for the
				// tokens created before September 2024. The lengths are not documented.
				"Notion API Token",
				"Notion API Token",
				*regexp.MustCompile(`\bntn_\d{11}[A-Za-z0-9]{35}\b|\bsecret_[A-Za-z0-9]{43}\b`),
				[]string{},
				"?",
				[]string{"https://developers.notion.com/page/changelog", "https://developers.notion.com/docs/authorization"},
			},
			{
				// nrt_, 11 digits and 34 or 35 characters: the length is the one of the
				// documented example.
				"Notion OAuth Refresh Token",
				"Notion OAuth Refresh Token",
				*regexp.MustCompile(`\bnrt_\d{11}[A-Za-z0-9]{34,35}\b`),
				[]string{},
				"?",
				[]string{"https://developers.notion.com/docs/authorization"},
			},
			{
				// oy2 and 43 base32 characters: an ID (v4 and 10 random bytes, 20 characters,
				// the first ones oy2) and a GUID (26 characters). The 4th, the 20th and the
				// last character also carry bits fixed by the prefix or by the padding. Keys
				// are issued in lowercase and checked in any case.
				"NuGet API Key",
				"NuGet API Key",
				*regexp.MustCompile(`(?i)\boy2[a-p][a-z2-7]{15}[aq][a-z2-7]{25}[aeimquy4]\b`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/nuget/nuget-org/publish-a-package",
					"https://github.com/NuGet/NuGetGallery/blob/main/src/NuGetGallery.Services/Authentication/ApiKeyV4.cs",
				},
			},
			{
				// nvapi- and 64 characters (the length is not documented), or the older NGC
				// keys (84 letters and digits) assigned to an NGC name, as NGC_API_KEY.
				"NVIDIA API Key",
				"NVIDIA API Key",
				*regexp.MustCompile(`\bnvapi-[\w-]{64}(?:[^\w-]|$)|` + keyed(`(?:\b|_)ngc`, ``, `[A-Za-z0-9]{84}\b`)),
				[]string{},
				"?",
				[]string{"https://docs.nvidia.com/nemo/retriever/latest/extraction/api-keys/"},
			},
			{
				// 00 and 40 characters, after the SSWS authentication scheme or assigned to
				// any Okta name: the IDs of Okta have 20 characters.
				"Okta API Token",
				"Okta API Token",
				*regexp.MustCompile(`\bSSWS\s{1,5}00[\w-]{40}(?:[^\w-]|$)|` + keyed(`okta`, ``, `00[\w-]{40}(?:[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{
					"https://developer.okta.com/docs/guides/create-an-api-token/main/",
					"https://devforum.okta.com/t/api-token-length/5519",
				},
			},
			{
				// 40 to 64 characters assigned to an Okta secret name (the length is not
				// documented). The client IDs are info.
				"Okta Client Secret",
				"Okta Client Secret",
				*regexp.MustCompile(keyed(`okta`, `secret`, `[\w-]{40,64}(?:[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{"https://developer.okta.com/docs/guides/implement-grant-type/authcode/main/"},
			},
			{
				"OpenAI API key",
				"OpenAI API key",
				// Project, service account, admin and project-less keys, or the
				// older ones with "OpenAI" in base64 (T3BlbkFJ) in the middle.
				*regexp.MustCompile(`\bsk-(?:(?:admin|None|proj|svcacct)-[\w-]{40,}|[A-Za-z0-9]{20}T3BlbkFJ[A-Za-z0-9]{20})`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// sk-or-v1- and 64 hex digits, also for the management keys.
				"OpenRouter API Key",
				"OpenRouter API Key",
				*regexp.MustCompile(`\bsk-or-v1-[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://openrouter.ai/docs/api/reference/authentication"},
			},
			{
				// REST API keys (20 characters) and Events API integration keys (32) assigned
				// to any PagerDuty name, or a REST API key in the Authorization header (Token
				// token=) on the same line as api.pagerduty.com: other APIs use that header.
				"PagerDuty API Key",
				"PagerDuty API Key",
				*regexp.MustCompile(keyed(`pager[_-]?duty`, ``, `(?:[\w+-]{20}|[A-Za-z0-9]{32})(?:[^\w+-]|$)`) +
					`|(?i:\bapi\.pagerduty\.com\b)[^\n]{0,200}?Token token=["']?[\w+-]{20}(?:[^\w+-]|$)` +
					`|Token token=["']?[\w+-]{20}["']?[^\w+\n-][^\n]{0,200}?(?i:\bapi\.pagerduty\.com\b)`),
				[]string{},
				"?",
				[]string{
					"https://developer.pagerduty.com/docs/authentication",
					"https://support.pagerduty.com/main/docs/api-access-keys",
				},
			},
			{
				"PayPal Braintree access token",
				"PayPal Braintree access token",
				*regexp.MustCompile(`access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Client secrets (E and 59 to 79 characters) assigned to any PayPal name:
				// the client IDs start with A.
				"PayPal Client Secret",
				"PayPal Client Secret",
				*regexp.MustCompile(keyed(`paypal`, ``, `E[\w-]{59,79}(?:[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{
					"https://developer.paypal.com/api/rest/",
					"https://github.com/paypal/PayPal-PHP-SDK/blob/master/sample/bootstrap.php",
				},
			},
			{
				// OAuth 2.0 access tokens, valid for some hours.
				"PayPal Access Token",
				"PayPal Access Token",
				*regexp.MustCompile(`\bA21AA[\w-]{60,200}`),
				[]string{},
				"?",
				[]string{"https://developer.paypal.com/api/rest/authentication/"},
			},
			{
				"Perplexity API Key",
				"Perplexity API Key",
				*regexp.MustCompile(`\bpplx-[A-Za-z0-9]{48}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// pcsk_ (or pckey_, the documented format of new keys), a public label, _ and
				// the secret: the lengths are not documented.
				"Pinecone API Key",
				"Pinecone API Key",
				*regexp.MustCompile(`\bpc(?:sk|key)_[A-Za-z0-9]{4,10}_[A-Za-z0-9]{40,100}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.pinecone.io/reference/api/2025-10/admin/create_api_key",
					"https://docs.pinecone.io/reference/cli/command-reference",
				},
			},
			{
				// access-, the environment (sandbox or production) and a UUID: the development
				// environment was retired.
				"Plaid Access Token",
				"Plaid Access Token",
				*regexp.MustCompile(`\baccess-(?:sandbox|production)-` + uuid + `\b`),
				[]string{},
				"?",
				[]string{"https://plaid.com/docs/api/items/", "https://plaid.com/docs/quickstart/glossary/"},
			},
			{
				// 30 characters assigned to any Plaid name: the client ID has 24.
				"Plaid Secret",
				"Plaid Secret",
				*regexp.MustCompile(keyed(`plaid`, ``, `[a-z0-9]{30}\b`)),
				[]string{},
				"?",
				[]string{"https://plaid.com/docs/quickstart/glossary/"},
			},
			{
				// With a user, a password and a host (see connFalsePositive): postgres://,
				// postgresql:// and the postgresql+driver:// of SQLAlchemy.
				"PostgreSQL Connection String",
				"PostgreSQL Connection String",
				*regexp.MustCompile(connectionString(`postgres(?:ql)?(?:\+\w{1,20})?`)),
				[]string{"example", "xxxx", "%2a%2a"},
				"?",
				[]string{
					"https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING",
					"https://docs.sqlalchemy.org/en/20/core/engines.html",
				},
			},
			{
				"Postman API Key",
				"Postman API Key",
				*regexp.MustCompile(`\bPMAK-[0-9a-f]{24}-[0-9a-f]{34}\b`),
				[]string{},
				"?",
				[]string{"https://learning.postman.com/docs/developer/postman-api/authentication/"},
			},
			{
				// Server and account API tokens (UUIDs) assigned to a Postmark key, token,
				// secret or password name, as POSTMARK_SERVER_TOKEN or the
				// X-Postmark-Server-Token header (message IDs are UUIDs too), or given to a
				// client of an official library, as new postmark.ServerClient("token").
				"Postmark API Token",
				"Postmark API Token",
				*regexp.MustCompile(keyed(`postmark`, secretWords, uuid+`\b`) +
					`|(?i:\bpostmark\.(?:server|account)client|\bpostmark(?:admin)?client|\bpostmark::(?:account)?apiclient\.new` +
					`|\bpostmark\.get(?:account)?apiclient)\(\s{0,5}["']` + uuid + `\b`),
				[]string{},
				"?",
				[]string{
					"https://postmarkapp.com/developer/api/overview",
					"https://postmarkapp.com/developer/api/messages-api",
					"https://github.com/ActiveCampaign/postmark-gem",
				},
			},
			{
				// pul- and 40 hex digits (the length is not documented): personal,
				// organization and team access tokens.
				"Pulumi Access Token",
				"Pulumi Access Token",
				*regexp.MustCompile(`\bpul-[0-9a-f]{40}\b`),
				[]string{},
				"?",
				[]string{
					"https://www.pulumi.com/blog/github-token-scanning-service/",
					"https://www.pulumi.com/docs/iac/cli/environment-variables/",
				},
			},
			{
				// The password of a connection in the text form of a Python Redis client or
				// connection pool, which lists every connection argument (e.g. in a log or
				// a debug page): None is not a password.
				"Python Redis Client Password",
				"Python Redis Client Password",
				*regexp.MustCompile(`\bredis\.(?:asyncio\.)?connection\.\w{1,40}\([^\n]{0,500}?\bpassword=[^,()\s"']{1,128}[,)]`),
				[]string{"password=none,", "password=none)"},
				"?",
				[]string{
					"https://redis.readthedocs.io/en/stable/connections.html",
					"https://github.com/redis/redis-py/blob/v6.4.0/redis/connection.py",
				},
			},
			{
				// amqp:// and amqps:// (also with a suffix, as in amqp+ssl://) with a user, a
				// password and a host (see connFalsePositive).
				"RabbitMQ Connection String",
				"RabbitMQ Connection String",
				*regexp.MustCompile(connectionString(`amqps?(?:\+\w{1,20})?`)),
				[]string{"example", "xxxx", "%2a%2a"},
				"?",
				[]string{"https://www.rabbitmq.com/docs/uri-spec"},
			},
			{
				// 50 characters (the length is not documented) assigned to any RapidAPI name,
				// as the X-RapidAPI-Key header: the X-RapidAPI-Host header is public.
				"RapidAPI Key",
				"RapidAPI Key",
				*regexp.MustCompile(keyed(`rapid[_-]?api`, ``, `[\w-]{50}(?:[^\w-]|$)`)),
				[]string{},
				"?",
				[]string{
					"https://docs.rapidapi.com/docs/configuring-api-security",
					"https://docs.rapidapi.com/docs/keys-and-key-rotation",
				},
			},
			{
				// redis:// and rediss:// (or with a suffix, as in redis+sentinel://) with a
				// password and a host, with or without a user (see connFalsePositive).
				"Redis Connection String",
				"Redis Connection String",
				*regexp.MustCompile(`(?i:rediss?(?:\+\w{1,20})?):\\?/\\?/(?:` + connUser + `)?:` + connPassword + `@` + connHost),
				[]string{"example", "xxxx", "%2a%2a"},
				"?",
				[]string{
					"https://redis.io/docs/latest/develop/clients/redis-py/connect/",
					"https://www.iana.org/assignments/uri-schemes/prov/redis",
				},
			},
			{
				// r8_ and 37 letters and digits.
				"Replicate API Token",
				"Replicate API Token",
				*regexp.MustCompile(`\br8_[A-Za-z0-9]{37}\b`),
				[]string{},
				"?",
				[]string{"https://replicate.com/docs/topics/security/api-tokens"},
			},
			{
				"Rubygems API Key",
				"Rubygems API Key",
				*regexp.MustCompile(`\brubygems_[0-9a-f]{48}\b`),
				[]string{},
				"?",
				[]string{
					"https://guides.rubygems.org/api-key-scopes/",
					"https://github.com/rubygems/rubygems.org/blob/master/app/controllers/concerns/api_keyable.rb",
				},
			},
			{
				// A session ID: the org ID (15 characters, 00D), ! (also escaped for a shell or
				// URL-encoded) and 96 characters. The length is not documented: 96 in the REST API
				// guide, whose 92-character placeholder is left out.
				"Salesforce Access Token",
				"Salesforce Access Token",
				*regexp.MustCompile(`\b00D[A-Za-z0-9]{12}(?:\\?!|%21)[\w.]{96}`),
				[]string{},
				"?",
				[]string{"https://developer.salesforce.com/docs/atlas.en-us.api_rest.meta/api_rest/quickstart_oauth.htm"},
			},
			{
				// Only the prefix is known, the format is not documented.
				"Salesforce Refresh Token",
				"Salesforce Refresh Token",
				*regexp.MustCompile(`\b5Aep861[\w.=]{80,251}`),
				[]string{},
				"?",
				[]string{"https://help.salesforce.com/s/articleView?id=xcloud.remoteaccess_oauth_refresh_token_flow.htm&language=en_US&type=5"},
			},
			{
				// 64 hex digits or 19 digits, assigned to a Salesforce secret name (replay IDs and
				// checksums look the same) or after the consumer key (3MVG9) as client_secret=.
				"Salesforce Consumer Secret",
				"Salesforce Consumer Secret",
				*regexp.MustCompile(keyed(`salesforce|sfdc|sf_client|sf_consumer`, `secret`, `(?:[0-9A-Fa-f]{64}|\d{19})\b`) +
					`|3MVG9[\w.+/=]{80,251}[^\n]{0,100}?client_secret=(?:[0-9A-Fa-f]{64}|\d{19})\b`),
				[]string{},
				"?",
				[]string{"https://help.salesforce.com/s/articleView?id=xcloud.remoteaccess_oauth_refresh_token_flow.htm&language=en_US&type=5"},
			},
			{
				// sgp_ and 64 letters and digits: a token of the Public API.
				"Segment Public API Token",
				"Segment Public API Token",
				*regexp.MustCompile(`\bsgp_[A-Za-z0-9]{64}\b`),
				[]string{},
				"?",
				[]string{
					"https://segment.com/docs/api/public-api/",
					"https://segment.com/blog/how-segment-proactively-protects-customer-api-tokens/",
				},
			},
			{
				// Two parts of 43 characters separated by a dot, as in the documentation,
				// assigned to any Segment name: a token of the Config API (deprecated).
				"Segment Config API Token",
				"Segment Config API Token",
				*regexp.MustCompile(keyed(`segment`, ``, `[\w-]{43}\.[\w-]{43}(?:[^\w.-]|$)`)),
				[]string{},
				"?",
				[]string{"https://segment.com/docs/api/config-api/authentication/"},
			},
			{
				// SG., the key ID and the secret (69 characters).
				"SendGrid API Key",
				"SendGrid API Key",
				*regexp.MustCompile(`\bSG\.[\w-]{22}\.[\w-]{43}`),
				[]string{},
				"?",
				[]string{"https://www.twilio.com/docs/sendgrid/ui/account-and-settings/api-keys"},
			},
			{
				// sntrys_, a payload (JSON in base64) and 43 characters.
				"Sentry Organization Auth Token",
				"Sentry Organization Auth Token",
				*regexp.MustCompile(`\bsntrys_eyJ[A-Za-z0-9+/=]{20,400}_[A-Za-z0-9+/]{43}(?:[^A-Za-z0-9+/=]|$)`),
				[]string{},
				"?",
				[]string{"https://docs.sentry.io/security-legal-pii/security/token-formats/"},
			},
			{
				// Personal tokens (sntryu_; sntrya_ and sntryi_ are reserved), or the older 64
				// hex digits (also internal integration tokens and client secrets) assigned to
				// a Sentry token, auth or secret name, to authToken within 200 characters
				// after "sentry" (e.g. in the build plugins), to auth.token or to token in the
				// auth section of the CLI configuration: the checksums of the CLI look the same.
				"Sentry Auth Token",
				"Sentry Auth Token",
				*regexp.MustCompile(`\bsntry[uai]_[0-9a-f]{64}\b|` + keyed(`sentry`, `token|auth|secret`, `[0-9a-f]{64}\b`) +
					`|(?i:sentry[\s\S]{0,200}?auth_?token)` + assign + `[0-9a-f]{64}\b` +
					`|(?:\bauth\.token|\[auth\]\s{1,5}token)\s{0,5}=\s{0,5}[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.sentry.io/security-legal-pii/security/token-formats/",
					"https://docs.sentry.io/cli/configuration/",
				},
			},
			{
				// Incoming webhooks, workflow webhooks and workflow triggers, also with
				// JSON-escaped slashes.
				"Slack Webhook",
				"Slack Webhook",
				*regexp.MustCompile(`https:\\?/\\?/hooks\.slack\.com\\?/(?:(?:services\\?/)?[TE][A-Z0-9]{8,12}\\?/B[A-Z0-9]{8,12}` +
					`\\?/[A-Za-z0-9]{24}|workflows\\?/[TE][A-Z0-9]{8,12}\\?/A[A-Z0-9]{8,12}\\?/\d{8,20}\\?/[A-Za-z0-9]{20,}|` +
					`triggers\\?/[TE][A-Z0-9]{8,12}\\?/\d{8,20}\\?/[0-9a-f]{20,40})`),
				[]string{},
				"?",
				[]string{
					"https://docs.slack.dev/messaging/sending-messages-using-incoming-webhooks",
					"https://slack.com/help/articles/360041352714-Create-more-advanced-workflows-using-webhooks",
				},
			},
			{
				// Personal access and service account tokens: snyk_uat. or snyk_sat., an ID (8
				// characters), a payload and a signature; or the older API tokens (UUIDs)
				// assigned to a Snyk key, token or secret name (the organization and project
				// IDs are UUIDs too) or given to snyk auth or snyk config set api=.
				"Snyk API Token",
				"Snyk API Token",
				*regexp.MustCompile(`\bsnyk_[us]at\.[a-z0-9]{8}\.[\w-]{8,1000}\.[\w-]{8,1000}(?:[^\w.-]|$)|` +
					keyed(`snyk`, secretWords, uuid+`\b`) + `|(?i:\bsnyk\s{1,5}(?:auth|config\s{1,5}set\s{1,5}api=)\s{0,5})` + uuid + `\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.snyk.io/developer-tools/snyk-cli/authenticate-to-use-the-cli",
					"https://github.com/snyk/go-application-framework/blob/main/pkg/auth/tokenauthenticator.go",
				},
			},
			{
				// User (squ_), global analysis (sqa_) and project analysis (sqp_) tokens: sq,
				// the type, _ and 40 hex digits; SonarQube Cloud organization tokens (sqco_,
				// the length is not documented); or the older 40 hex digits assigned to a
				// Sonar token, login, key or secret name: commit hashes look the same. Badge
				// tokens (sqb_) only show the status of a project.
				"SonarQube Token",
				"SonarQube Token",
				*regexp.MustCompile(`\bsq[apu]_[0-9a-f]{40}\b|\bsqco_[A-Za-z0-9]{40,80}\b|` +
					keyed(`sonar`, `login|`+secretWords, `[0-9a-f]{40}\b`)),
				[]string{},
				"?",
				[]string{
					"https://docs.sonarsource.com/sonarqube-server/latest/user-guide/managing-tokens/",
					"https://github.com/SonarSource/sonarqube/blob/master/server/sonar-webserver-auth/src/main/java/org/sonar/server/usertoken/TokenGeneratorImpl.java",
					"https://docs.sonarsource.com/sonarqube-cloud/administering-sonarcloud/managing-organization/scoped-organization-tokens",
				},
			},
			{
				// Organization access tokens of Splunk Observability Cloud (formerly SignalFx):
				// 22 characters (the format is not documented) assigned to a SignalFx or
				// Splunk Observability token name (also in a Helm value, in a block of a
				// collector or in the Terraform provider), to SPLUNK_ACCESS_TOKEN or
				// SPLUNK_HEC_TOKEN, sent in the X-SF-Token header or given to a client of
				// the Node.js library (signalfx.Ingest, signalfx.SignalFlow). RUM tokens
				// (rumAccessToken) are public: they ship with the page.
				"Splunk Observability Access Token",
				"Splunk Observability Access Token",
				*regexp.MustCompile(keyed(`signalfx|splunk[_.-]?(?:observability|hec)`, `token`, `[\w-]{22}(?:[^\w-]|$)`) + `|` +
					keyed(`splunk[_.-]?access[_.-]?token|x-sf-token`, ``, `[\w-]{22}(?:[^\w-]|$)`) +
					`|(?i:\bsignalfx\.(?:ingest(?:json)?|signalflow)\(\s{0,5})["'\x60][\w-]{22}(?:[^\w-]|$)` +
					`|(?i:signalfx"\s{0,5}\{[^}]{0,200}?auth_token\s{0,5}=\s{0,5}")[\w-]{22}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://help.splunk.com/en/splunk-observability-cloud/administer/authentication-and-security/authentication-tokens/org-access-tokens",
					"https://help.splunk.com/en/splunk-observability-cloud/manage-data/splunk-distribution-of-the-opentelemetry-collector/get-started-with-the-splunk-distribution-of-the-opentelemetry-collector/get-started-understand-and-use-the-collector/environment-variables",
					"https://help.splunk.com/en/splunk-observability-cloud/manage-data/other-data-ingestion-methods/other-data-ingestion-methods/send-data-using-rest-apis",
					"https://help.splunk.com/en/splunk-observability-cloud/digital-experience-monitoring/real-user-monitoring/visibility-of-rum-tokens",
					"https://github.com/signalfx/splunk-otel-collector-chart",
					"https://registry.terraform.io/providers/splunk-terraform/signalfx/latest/docs",
					"https://github.com/signalfx/signalfx-nodejs",
				},
			},
			{
				// Secret and restricted keys, live mode, and organization keys.
				"Stripe API key",
				"Stripe API key",
				*regexp.MustCompile(`\b(?:[rs]k_live|sk_org)_[0-9a-zA-Z_]{24,247}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.stripe.com/keys",
					"https://docs.stripe.com/keys/organization-api-keys",
				},
			},
			{
				"Stripe Test API key",
				"Stripe Test API key",
				*regexp.MustCompile(`\b[rs]k_test_[0-9a-zA-Z]{24,247}\b`),
				[]string{},
				"?",
				[]string{"https://docs.stripe.com/keys"},
			},
			{
				"Square access token",
				"Square access token",
				*regexp.MustCompile(`\b(?:sq0atp-[\w-]{22}|EAAA[\w-]{60})`),
				[]string{},
				"?",
				[]string{"https://developer.squareup.com/docs/build-basics/access-tokens"},
			},
			{
				// The application secret, also of the Sandbox.
				"Square OAuth secret",
				"Square OAuth secret",
				*regexp.MustCompile(`\b(?:sandbox-)?sq0cs[bp]-[\w-]{43}`),
				[]string{},
				"?",
				[]string{"https://developer.squareup.com/docs/build-basics/access-tokens"},
			},
			{
				// sb_secret_ and 31 characters (the length is not documented): the
				// publishable keys are info, the older keys are JSON Web Tokens.
				"Supabase Secret Key",
				"Supabase Secret Key",
				*regexp.MustCompile(`\bsb_secret_[\w-]{31}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{"https://supabase.com/docs/guides/api/api-keys"},
			},
			{
				// Personal access tokens of the Management API: sbp_ and 40 hex digits.
				"Supabase Access Token",
				"Supabase Access Token",
				*regexp.MustCompile(`\bsbp_[0-9a-f]{40}\b`),
				[]string{},
				"?",
				[]string{"https://supabase.com/docs/reference/api/introduction"},
			},
			{
				// The bot ID, : and AA with 32-33 characters (AA and the length are not documented,
				// the examples have them), also after bot as in the API URLs.
				"Telegram Bot Token",
				"Telegram Bot Token",
				*regexp.MustCompile(`(?:\b|bot)\d{5,16}:AA[\w-]{32,33}(?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://core.telegram.org/bots/api",
					"https://core.telegram.org/bots/features",
				},
			},
			{
				// API tokens: 16 random bytes in base64url (22 characters, the last one A, Q,
				// g or w) assigned to a Travis key, token, secret or password name (the
				// default environment variables hold public values of the same shape, as
				// TRAVIS_BRANCH), or in the Authorization header (token) on the same line as
				// and after api.travis-ci.com or the Travis-API-Version header: other APIs use
				// that header. The encrypted variables (secure:) are public by design.
				"Travis CI API Token",
				"Travis CI API Token",
				*regexp.MustCompile(keyed(`travis`, secretWords, `[\w-]{21}[AQgw](?:[^\w-]|$)`) +
					`|(?i:\bapi\.travis-ci\.com\b|\btravis-api-version\b)[^\n]{0,200}?(?i:authorization)` + assign + `token [\w-]{21}[AQgw](?:[^\w-]|$)`),
				[]string{},
				"?",
				[]string{
					"https://developer.travis-ci.com/authentication",
					"https://github.com/travis-ci/travis-api/blob/master/lib/travis/api/app/access_token.rb",
					"https://docs.travis-ci.com/user/environment-variables/",
					"https://docs.travis-ci.com/user/encryption-keys/",
				},
			},
			{
				// ATTA and 72 letters and digits (the current format, not documented), or the
				// older 64 hex digits assigned to any Trello name or in the token parameter of
				// an API URL: the API key is public, the token is not.
				"Trello Token",
				"Trello Token",
				*regexp.MustCompile(`\bATTA[A-Za-z0-9]{72}\b|` + keyed(`trello`, ``, `[0-9a-f]{64}\b`) +
					`|\btrello\.com/1/[^\s"'<>]{0,200}?[?&]token=[0-9a-f]{64}\b`),
				[]string{},
				"?",
				[]string{"https://developer.atlassian.com/cloud/trello/guides/rest-api/authorization/"},
			},
			{
				// Auth tokens (32 hex digits) and API key secrets (32 characters) assigned to
				// any Twilio name, or with their account or API key SID (the user of Basic
				// authentication): assigned to an auth token or secret name after it, or
				// within 300 characters after "twilio" as SID:secret or "SID", "secret" (as
				// in a client). The SIDs (34 characters) are info.
				"Twilio Secret",
				"Twilio Secret",
				*regexp.MustCompile(keyed(`twilio`, ``, `[A-Za-z0-9]{32}\b`) + `|\b(?:AC|SK)[0-9a-f]{32}\b[\s\S]{0,100}?(?i:auth_?token|secret)` +
					`[\w.-]{0,10}` + assign + `[A-Za-z0-9]{32}\b|(?i:twilio)[\s\S]{0,300}?\b(?:AC|SK)[0-9a-f]{32}` +
					`(?::|["'\x60]\s{0,5},\s{0,5}["'\x60])[A-Za-z0-9]{32}\b`),
				[]string{},
				"?",
				[]string{
					"https://www.twilio.com/docs/iam/api-keys/api-key",
					"https://www.twilio.com/docs/glossary/what-is-a-sid",
					"https://www.twilio.com/docs/usage/requests-to-twilio",
				},
			},
			{
				// The prefix, the public part (24 characters) and the secret (64).
				"Dynatrace token",
				"Dynatrace token",
				*regexp.MustCompile(`\bdt0[a-zA-Z]\d{2}\.[A-Z0-9]{24}\.[A-Z0-9]{64}\b`),
				[]string{},
				"?",
				[]string{
					"https://docs.dynatrace.com/docs/manage/identity-access-management/access-tokens-and-oauth-clients/access-tokens-legacy",
				},
			},
			{
				"Shopify shared secret",
				"Shopify shared secret",
				*regexp.MustCompile(`\bshpss_[a-fA-F0-9]{32}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				"Shopify access token",
				"Shopify access token",
				*regexp.MustCompile(`\bshpat_[a-fA-F0-9]{32}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				"Shopify custom app access token",
				"Shopify custom app access token",
				*regexp.MustCompile(`\bshpca_[a-fA-F0-9]{32}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				"Shopify private app access token",
				"Shopify private app access token",
				*regexp.MustCompile(`\bshppa_[a-fA-F0-9]{32}\b`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// pypi- and a macaroon for pypi.org (base64).
				"PyPI upload token",
				"PyPI upload token",
				*regexp.MustCompile(`\bpypi-AgEIcHlwaS5vcmc[\w-]{50,1000}`),
				[]string{},
				"?",
				[]string{"https://docs.pypi.org/api/secrets/"},
			},
			{
				// The webhook ID and token, also with JSON-escaped slashes.
				"Discord Webhook",
				"Discord Webhook",
				*regexp.MustCompile(`https:\\?/\\?/(?:(?:ptb|canary)\.)?discord(?:app)?\.com\\?/api(?:\\?/v\d{1,2})?\\?/webhooks\\?/\d{17,20}\\?/[\w-]{60,72}`),
				[]string{},
				"?",
				[]string{"https://discord.com/developers/docs/resources/webhook"},
			},
			{
				"Google OAuth Access Key",
				"Google OAuth Access Key",
				*regexp.MustCompile(`\bya29\.[\w-]{30,}`),
				[]string{},
				"?",
				[]string{},
			},
			{
				// Secret (sk) and temporary (tk) tokens: the public ones (pk) are info.
				"Mapbox Token",
				"Mapbox Token",
				*regexp.MustCompile(`\b[st]k\.eyJ1Ijoi[\w-]{10,}\.[\w-]{20,}`),
				[]string{},
				"?",
				[]string{"https://docs.mapbox.com/api/accounts/tokens/"},
			},
			{
				// Incoming webhooks, also the older outlook.office.com ones.
				"Microsoft Teams Webhook",
				"Microsoft Teams Webhook",
				*regexp.MustCompile(`https://(?:outlook\.office\.com/webhook|[\w-]+\.webhook\.office\.com/webhookb2)/[\w@-]{20,100}/IncomingWebhook/` +
					`[\w-]{20,64}/[\w-]{20,64}(?:/[\w-]{20,64})?`),
				[]string{},
				"?",
				[]string{
					"https://learn.microsoft.com/en-us/microsoftteams/platform/webhooks-and-connectors/how-to/add-incoming-webhook",
				},
			},
			{
				// Workflows started by an HTTP request (Teams webhooks, Logic Apps): sig signs
				// the URL.
				"Microsoft Power Automate Workflow URL",
				"Microsoft Power Automate Workflow URL",
				*regexp.MustCompile(`https://[\w.-]{1,100}\.(?:logic\.azure\.com|powerplatform\.com)(?::443)?/[^\s"'<>?]{0,200}/triggers/` +
					`[^\s"'<>?/]{1,64}/paths/invoke\?[^\s"'<>]{0,300}?\bsig=[\w%-]{40,}`),
				[]string{},
				"?",
				[]string{
					"https://support.microsoft.com/en-us/office/create-incoming-webhooks-with-workflows-for-microsoft-teams-8ae491c7-0394-4861-ba59-055e33f75498",
				},
			},
			{
				// Personal (vcp_), integration (vci_), app access (vca_) and refresh (vcr_)
				// tokens and API keys (vck_): the length is not documented, 56 characters in
				// the examples. Or the older 24 characters assigned to a Vercel token name:
				// the older team and user IDs look the same.
				"Vercel Access Token",
				"Vercel Access Token",
				*regexp.MustCompile(`\bvc[aikpr]_[A-Za-z0-9]{40,64}\b|` + keyed(`vercel`, `token`, `[A-Za-z0-9]{24}\b`)),
				[]string{},
				"?",
				[]string{
					"https://vercel.com/changelog/new-token-formats-and-secret-scanning",
					"https://vercel.com/docs/sign-in-with-vercel/tokens",
					"https://vercel.com/docs/rest-api/authentication/create-an-auth-token",
				},
			},
			{
				// xai- and 80 characters (the length is not documented).
				"xAI API Key",
				"xAI API Key",
				*regexp.MustCompile(`\bxai-[A-Za-z0-9]{80}\b`),
				[]string{},
				"?",
				[]string{"https://docs.x.ai/docs/tutorial"},
			},
			{
				// Assigned to any Zendesk name.
				"Zendesk Secret Key",
				"Zendesk Secret Key",
				*regexp.MustCompile(keyed(`zendesk`, ``, `[A-Za-z0-9]{35,40}\b`)),
				[]string{},
				"?",
				[]string{"https://developer.zendesk.com/api-reference/introduction/security-and-auth/"},
			},
		}
	})

	return secrets
}

// RemoveDuplicateSecrets removes duplicates from secrets found.
func RemoveDuplicateSecrets(input []SecretMatched) []SecretMatched {
	keys := make(map[string]bool)
	list := []SecretMatched{}

	for _, entry := range input {
		if _, value := keys[entry.Match]; !value {
			keys[entry.Match] = true
			list = append(list, entry)
		}
	}

	return list
}

// GetSecretScanner returns the scanner for the predefined secret regexes.
func GetSecretScanner() *SecretScanner {
	onceSecretScanner.Do(func() {
		secretScanner = NewSecretScanner(GetSecretRegexes())
	})

	return secretScanner
}

// NewSecretScanner returns a scanner for the given secrets.
func NewSecretScanner(secrets []Secret) *SecretScanner {
	s := &SecretScanner{secrets: slices.Clone(secrets)}
	regexes := make([]*regexp.Regexp, len(s.secrets))

	for i := range s.secrets {
		regexes[i] = &s.secrets[i].Regex
	}

	s.engine = NewEngine(regexes)

	return s
}

// Scan returns the secrets found in body, in the order of the secrets
// list, excluding known false positives.
func (s *SecretScanner) Scan(target, body string) []SecretMatched {
	var results []SecretMatched

	for i, locs := range s.engine.FindAllIndex(body) {
		for _, loc := range locs {
			match := body[loc[0]:loc[1]]
			if containsFalsePositive(match, s.secrets[i].FalsePositives) || uriFalsePositive(s.secrets[i].Name, match) {
				continue
			}

			results = append(results, SecretMatched{Secret: s.secrets[i], URL: target, Match: match})
		}
	}

	return results
}

// uriFalsePositive reports whether the match of the rule name is a URI that
// doesn't hold real credentials, because of its host or password.
func uriFalsePositive(name, match string) bool {
	switch name {
	case "Credentials in URL":
		return exampleHost(uriHost(match)) || ownScheme(match)
	case "MongoDB Connection String", "MySQL Connection String", "PostgreSQL Connection String", "RabbitMQ Connection String",
		"Redis Connection String":
		return connFalsePositive(match)
	case "JDBC Connection String":
		return jdbcFalsePositive(match)
	case "Database Connection String":
		return propFalsePositive(match)
	}

	return false
}

// ownScheme reports whether the URL match has a scheme with a rule of its own,
// in any case: Credentials in URL leaves it to that rule. Like those rules, a
// name matches at the end of a longer word (\nmongodb, odbc+postgres) and with
// a suffix after + (mongodb+srv, postgresql+asyncpg). A new rule for the URLs
// of a scheme must add the scheme here.
func ownScheme(match string) bool {
	for part := range strings.SplitSeq(strings.ToLower(match[:strings.IndexByte(match, ':')]), "+") {
		for _, scheme := range []string{"mongodb", "mysql", "mysqlx", "mysql2", "postgres", "postgresql", "redis", "rediss", "cloudinary", "amqp", "amqps"} {
			if strings.HasSuffix(part, scheme) {
				return true
			}
		}
	}

	return false
}

// uriHost returns the host of the URI match (...@host[:port]...), in lowercase.
func uriHost(match string) string {
	host := strings.ToLower(match[strings.LastIndexByte(match, '@')+1:])
	if strings.HasPrefix(host, "[") {
		return host[:strings.IndexByte(host, ']')+1]
	}

	end := strings.IndexFunc(host, func(r rune) bool {
		return !strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789.-", r)
	})
	if end >= 0 {
		return host[:end]
	}

	return host
}

// exampleHost reports whether host is reserved for documentation (RFC 2606):
// example.com, example.net, example.org or one of their subdomains, or a name
// in .example, .test or .invalid.
func exampleHost(host string) bool {
	dot := strings.LastIndexByte(host, '.')
	if dot < 0 {
		return false // a single label, e.g. localhost or test.
	}

	switch host[dot+1:] {
	case "example", "test", "invalid":
		return true
	case "com", "net", "org":
		return host[:dot] == "example" || strings.HasSuffix(host[:dot], ".example")
	}

	return false
}

// localHost reports whether host is the local machine (localhost, a loopback
// or unspecified address) or an address reserved for documentation (RFC 5737,
// RFC 3849).
func localHost(host string) bool {
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}

	addr, err := netip.ParseAddr(strings.Trim(host, "[]"))
	if err != nil {
		return false
	}

	return addr.IsLoopback() || addr.IsUnspecified() ||
		slices.ContainsFunc(documentationAddresses, func(p netip.Prefix) bool { return p.Contains(addr) })
}

// connFalsePositive reports whether the connection URI match
// (scheme://user:password@host...) is not a real one: its host is local or
// reserved for documentation.
func connFalsePositive(match string) bool {
	host := uriHost(match)

	return exampleHost(host) || localHost(host)
}

// jdbcFalsePositive reports whether the JDBC URL match is not a real one: its
// host (after // or, in an Oracle URL, after @) is local or reserved for
// documentation. A URL without a host (e.g. of an embedded database) is real.
func jdbcFalsePositive(match string) bool {
	url := strings.ToLower(strings.ReplaceAll(match, `\/`, "/"))

	_, host, found := strings.Cut(url, "//")
	if !found {
		if _, host, found = strings.Cut(url, "@"); !found {
			return false
		}
	}

	if strings.HasPrefix(host, "[") {
		host = host[:strings.IndexByte(host, ']')+1]
	} else if end := strings.IndexAny(host, ":/;?,\\"); end >= 0 {
		host = host[:end]
	}

	return exampleHost(host) || localHost(host)
}

// propServerValue matches the server of a connection string made of
// properties: the protocol, the host, and the port or the instance.
var propServerValue = regexp.MustCompile(`(?i)\b(?:data source|server|host)\s{0,3}=\s{0,3}(?:[a-z]{2,5}:)?([^;,:\\\s]*)`)

// propFalsePositive reports whether the connection string match (properties)
// is not a real one: its server is the local machine (localhost, a loopback
// address, ".", (local) or a LocalDB instance) or reserved for documentation.
func propFalsePositive(match string) bool {
	server := propServerValue.FindStringSubmatch(match)
	if server == nil {
		return false
	}

	host := strings.ToLower(server[1])

	return host == "." || host == "(local)" || host == "(localdb)" || exampleHost(host) || localHost(host)
}

// containsFalsePositive reports whether match contains, ignoring case,
// one of the (lowercase) false positives.
func containsFalsePositive(match string, falsePositives []string) bool {
	if len(falsePositives) == 0 {
		return false
	}

	match = strings.ToLower(match)
	for _, fp := range falsePositives {
		if strings.Contains(match, fp) {
			return true
		}
	}

	return false
}
