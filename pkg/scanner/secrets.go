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

package scanner

// Secret struct.
// Name = the name that identifies the secret.
// Description.
// Regex = The regular expression matching the secret.
// FalsePositives = A list of known false positives.
// PoC = cli command to check if the secret is valid or not.
type Secret struct {
	Name           string
	Description    string
	Regex          string
	FalsePositives []string
	Poc            string
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

// GetSecretRegexes returns a slice of all
// the secret structs.
func GetSecretRegexes() []Secret {
	var regexes = []Secret{
		{
			"Adafruit API Key",
			"Adafruit API Key",
			`(?i)(?:adafruit)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}([a-z0-9_-]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Adobe Client ID",
			"Adobe Client ID",
			`(?i)(?:adobe)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}(?:=|>|:{1,3}` +
				`=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}([a-f0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Adobe Client Secret",
			"Adobe Client Secret",
			`(?i)\b((p8e-)(?i)[a-z0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Age Secret Key",
			"Age Secret Key",
			`(?i)AGE-SECRET-KEY-1[QPZRY9X8GF2TVDW0S3JN54KHCE6MUA7L]{58}`,
			[]string{},
			"?",
		},
		{
			"Airtable API Key",
			"Airtable API Key",
			`(?i)(?:airtable)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{17})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Algolia API Key",
			"Algolia API Key",
			`(?i)(?:algolia)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Alibaba Access Key ID",
			"Alibaba Access Key ID",
			`(?i)\b((LTAI)(?i)[a-z0-9]{20})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Alibaba Secret Key",
			"Alibaba Secret Key",
			`(?i)(?:alibaba)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{30})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Asana Client ID",
			"Asana Client ID",
			`(?i)(?:asana)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([0-9]{16})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Asana Client Secret",
			"Asana Client Secret",
			`(?i)(?:asana)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Atlassian API Token",
			"Atlassian API Token",
			`(?i)(?:atlassian|confluence|jira)(?:[0-9a-z\-_\t .]{0,20})` +
				`(?:[\s|']|[\s|"]){0,3}(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{24})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"AWS Access Key",
			"AWS Access Key",
			"(A3T[A-Z0-9]|AKIA|ACCA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA|ASCA|APKA)[A-Z0-9]{16}",
			[]string{},
			"?",
		},
		{
			"AWS Secret Key",
			"AWS Secret Key",
			`(?i)aws(.{0,20})?(?-i)['\"][0-9a-zA-Z\/+]{40}['\"]`,
			[]string{},
			"?",
		},
		{
			"AWS MWS Key",
			"AWS MWS Key",
			`amzn\.mws\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
			[]string{},
			"?",
		},
		{
			"Amazon SNS topic",
			"Amazon SNS topic",
			`arn:aws:sns:[a-z0-9\-]+:[0-9]+:[A-Za-z0-9\-_]+`,
			[]string{},
			"?",
		},
		{
			"Beamer API Token",
			"Beamer API Token",
			`(?i)(?:beamer)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`(b_[a-z0-9=_\-]{44})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"BitBucket Client ID",
			"BitBucket Client ID",
			`(?i)(?:bitbucket)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"BitBucket Client Secret",
			"BitBucket Client Secret",
			`(?i)(?:bitbucket)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9=_\-]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Cloudflare API Key",
			"Cloudflare API Key",
			`(?i)(?:cloudflare)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9_-]{40})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Cloudflare Global API Key",
			"Cloudflare Global API Key",
			`(?i)(?:cloudflare)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-f0-9]{37})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Cloudflare Origin CA Key",
			"Cloudflare Origin CA Key",
			`\b(v1\.0-[a-f0-9]{24}-[a-f0-9]{146})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"CodeCov Access Token",
			"CodeCov Access Token",
			`(?i)(?:codecov)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"CoinBase Access Token",
			"CoinBase Access Token",
			`(?i)(?:coinbase)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9_-]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Confluent Access Token",
			"Confluent Access Token",
			`(?i)(?:confluent)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{16})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Confluent Secret Key",
			"Confluent Secret Key",
			`(?i)(?:confluent)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Databricks API Token",
			"Databricks API Token",
			`(?i)\b(dapi[a-h0-9]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"DataDog Access Token",
			"DataDog Access Token",
			`(?i)(?:datadog)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{40})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"DigitalOcean Access Token",
			"DigitalOcean Access Token",
			`(?i)\b(doo_v1_[a-f0-9]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"DigitalOcean Personal Access Token",
			"DigitalOcean Personal Access Token",
			`(?i)\b(dop_v1_[a-f0-9]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"DigitalOcean Refresh Token",
			"DigitalOcean Refresh Token",
			`(?i)\b(dor_v1_[a-f0-9]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Discord API Token",
			"Discord API Token",
			`(?i)(?:discord)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-f0-9]{64})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Discord Client ID",
			"Discord Client ID",
			`(?i)(?:discord)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([0-9]{18})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Discord Client Secret",
			"Discord Client Secret",
			`(?i)(?:discord)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9=_\-]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"DropBox API Token",
			"DropBox API Token",
			`(?i)(?:dropbox)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{15})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
		{
			"Facebook Secret Key",
			"Facebook Secret Key",
			`(?i)(facebook|fb)(.{0,20})?(?-i)['\"][0-9a-f]{32}['\"]`,
			[]string{"facebook.com", "facebook.svg"},
			"?",
		},
		{
			"Facebook Client ID",
			"Facebook Client ID",
			`(?i)(facebook|fb)(.{0,20})?['\"][0-9]{13,17}['\"]`,
			[]string{"facebook.com", "facebook.svg"},
			"?",
		},
		{
			"Fastly API Token",
			"Fastly API Token",
			`(?i)(?:fastly)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9=_\-]{32})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{"facebook.com", "facebook.svg"},
			"?",
		},
		{
			"Cloudinary Basic Auth",
			"Cloudinary Basic Auth",
			`cloudinary://[0-9]{15}:[0-9A-Za-z\-_]+@[0-9A-Za-z\-_]+`,
			[]string{},
			"?",
		},
		{
			"Firebase Database",
			"Firebase Database",
			`([a-z0-9.-]+\.firebaseio\.com|[a-z0-9.-]+\.firebaseapp\.com)`,
			[]string{},
			"?",
		},
		{
			"Twitter Secret Key",
			"Twitter Secret Key",
			`(?i)twitter(.{0,20})?[0-9a-z]{35,44}`,
			[]string{"twitter.com"},
			"?",
		},
		{
			"Twitter Client ID",
			"Twitter Client ID",
			`(?i)twitter(.{0,20})?[0-9a-z]{18,25}`,
			[]string{"twitter.com"},
			"?",
		},
		{
			"Github Personal Access Token",
			"Github Personal Access Token",
			`ghp_.{36}`,
			[]string{},
			"?",
		},
		{
			"Github OAuth Access Token",
			"Github OAuth Access Token",
			`gho_.{36}`,
			[]string{},
			"?",
		},
		{
			"Github App Token",
			"Github App Token",
			`(ghu|ghs)_.{36}`,
			[]string{},
			"?",
		},
		{
			"Github Refresh Token",
			"Github Refresh Token",
			`ghr_.{76}`,
			[]string{},
			"?",
		},
		{
			"LinkedIn Client ID",
			"LinkedIn Client ID",
			`(?i)linkedin(.{0,20})?(?-i)[0-9a-z]{12}`,
			[]string{"linkedin.com", "linkedin.svg"},
			"?",
		},
		{
			"LinkedIn Secret Key",
			"LinkedIn Secret Key",
			`(?i)linkedin(.{0,20})?[0-9a-z]{16}`,
			[]string{"linkedin.com", "linkedin.svg"},
			"?",
		},
		{
			"Slack",
			"Slack",
			`xox[baprs]-([0-9a-zA-Z]{10,48})?`,
			[]string{},
			"?",
		},
		{
			"Asymmetric Private Key",
			"Asymmetric Private Key",
			`-----BEGIN ((EC|PGP|DSA|RSA|OPENSSH) )?PRIVATE KEY( BLOCK)?-----`,
			[]string{},
			"?",
		},
		{
			"Google API key",
			"Google API key",
			`AIza[0-9A-Za-z\-_]{35}`,
			[]string{},
			"?",
		},
		{
			"Google (GCP) Service Account",
			"Google (GCP) Service Account",
			`"type": "service_account"`,
			[]string{},
			"?",
		},
		{
			"Heroku API key",
			"Heroku API key",
			`(?i)heroku(.{0,20})?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
			[]string{},
			"?",
		},
		{
			"MailChimp API key",
			"MailChimp API key",
			`[0-9a-f]{32}-us[0-9]{1,2}`,
			[]string{},
			"?",
		},
		{
			"Mailgun API key",
			"Mailgun API key",
			`key\-[0-9a-zA-Z]{32}`,
			[]string{},
			"?",
		},
		{
			"PayPal Braintree access token",
			"PayPal Braintree access token",
			`access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}`,
			[]string{},
			"?",
		},
		{
			"Picatic API key",
			"Picatic API key",
			`sk\_live\_[0-9a-z]{32}`,
			[]string{},
			"?",
		},
		{
			"SendGrid API Key",
			"SendGrid API Key",
			`SG\.[a-zA-Z0-9]{22}\.[a-zA-Z0-9]{43}`,
			[]string{},
			"?",
		},
		{
			"Slack Webhook",
			"Slack Webhook",
			`https\:\/\/hooks\.slack\.com/services/T[0-9A-Za-z\-_]{8}/B[0-9A-Za-z\-_]{8}/[0-9A-Za-z\-_]{24}`,
			[]string{},
			"?",
		},
		{
			"Stripe API key",
			"Stripe API key",
			`(?i)stripe(.{0,20})?[sr]k_live_[0-9a-zA-Z]{24}`,
			[]string{},
			"?",
		},
		{
			"Square access token",
			"Square access token",
			`sq0atp\-[0-9A-Za-z\-_]{22}|EAAAE[a-zA-Z0-9\-_]{59}`,
			[]string{},
			"?",
		},
		{
			"Square OAuth secret",
			"Square OAuth secret",
			`sq0csp\-[0-9A-Za-z\-_]{43}`,
			[]string{},
			"?",
		},
		{
			"Twilio API key",
			"Twilio API key",
			`(?i)twilio(.{0,20})?SK[0-9a-f]{32}`,
			[]string{},
			"?",
		},
		{
			"Dynatrace token",
			"Dynatrace token",
			`dt0[a-zA-Z]{1}[0-9]{2}\.[A-Z0-9]{24}\.[A-Z0-9]{64}`,
			[]string{},
			"?",
		},
		{
			"Shopify shared secret",
			"Shopify shared secret",
			`shpss\_[a-fA-F0-9]{32}`,
			[]string{},
			"?",
		},
		{
			"Shopify access token",
			"Shopify access token",
			`shpat\_[a-fA-F0-9]{32}`,
			[]string{},
			"?",
		},
		{
			"Shopify custom app access token",
			"Shopify custom app access token",
			`shpca\_[a-fA-F0-9]{32}`,
			[]string{},
			"?",
		},
		{
			"Shopify private app access token",
			"Shopify private app access token",
			`shppa\_[a-fA-F0-9]{32}`,
			[]string{},
			"?",
		},
		{
			"PyPI upload token",
			"PyPI upload token",
			`pypi\-AgEIcHlwaS5vcmc[A-Za-z0-9-_]{50,1000}`,
			[]string{},
			"?",
		},
		{
			"Bugsnag API Key",
			"Bugsnag API Key",
			`(?i)(bs|bugsnag)(.{0,20})?[0-9a-f]{32}`,
			[]string{},
			"?",
		},
		{
			"AWS cognito pool",
			"AWS Cognito pool",
			`(us-east-1|us-east-2|us-west-1|us-west-2|sa-east-1):[0-9A-Za-z]{8}-[0-9A-Za-z]{4}` +
				`-[0-9A-Za-z]{4}-[0-9A-Za-z]{4}-[0-9A-Za-z]{12}`,
			[]string{},
			"?",
		},
		{
			"Discord Webhook",
			"Discord Webhook",
			`https\:\/\/discordapp\.com\/api\/webhooks\/[0-9]+/[A-Za-z0-9\-]+`,
			[]string{},
			"?",
		},
		{
			"Google Calendar URI",
			"Google Calendar URI",
			`https\:\/\/(.*)calendar\.google\.com\/calendar\/[0-9a-z\/]+\/embed\?src=[A-Za-z0-9%@&;=\-_\.\/]+`,
			[]string{},
			"?",
		},
		{
			"Google OAuth Access Key",
			"Google OAuth Access Key",
			`ya29\.[0-9A-Za-z\-_]+`,
			[]string{},
			"?",
		},
		{
			"Mapbox Token Disclosure",
			"Mapbox Token Disclosure",
			`(pk|sk)\.eyJ1Ijoi\w+\.[\w-]*`,
			[]string{},
			"?",
		},
		{
			"Microsoft Teams Webhook",
			"Microsoft Teams Webhook",
			`https\:\/\/outlook\.office\.com\/webhook\/[A-Za-z0-9\-@]+\/IncomingWebhook\/[A-Za-z0-9\-]+\/[A-Za-z0-9\-]+`,
			[]string{},
			"?",
		},
		{
			"Alibaba OSS Bucket",
			"Alibaba OSS Bucket",
			`(?:[a-zA-Z0-9-\.\_]+\.oss-[a-zA-Z0-9-\.\_]+\.aliyuncs\.com|` +
				`oss-[a-zA-Z0-9-\.\_]+\.aliyuncs\.com/[a-zA-Z0-9-\.\_]+)`,
			[]string{},
			"?",
		},
		{
			"Zendesk Secret Key",
			"Zendesk Secret Key",
			`(?i)(?:zendesk)(?:[0-9a-z\-_\t .]{0,20})(?:[\s|']|[\s|"]){0,3}` +
				`(?:=|>|:{1,3}=|\|\|:|<=|=>|:|\?=)(?:'|\"|\s|=|\x60){0,5}` +
				`([a-z0-9]{40})(?:['|\"|\n|\r|\s|\x60|;]|$)`,
			[]string{},
			"?",
		},
	}

	return regexes
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
