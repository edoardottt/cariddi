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
*/

package scanner

//Secret struct
type Secret struct {
	Name           string
	Description    string
	Regex          string
	FalsePositives []string
	Poc            string
}

//SecretMatched struct
type SecretMatched struct {
	Secret Secret
	Url    string
	Match  string
}

//GetRegexes returns all the regexes
func GetRegexes() []Secret {
	var regexes = []Secret{
		{
			"AWS Access Key",
			"AWS Access Key",
			"(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}",
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
			"Facebook Secret Key",
			"Facebook Secret Key",
			"(?i)(facebook|fb)(.{0,20})?(?-i)['\"][0-9a-f]{32}['\"]",
			[]string{"facebook.com/", "facebook.svg"},
			"?",
		},
		{
			"Facebook Client ID",
			"Facebook Client ID",
			"(?i)(facebook|fb)(.{0,20})?['\"][0-9]{13,17}['\"]",
			[]string{"facebook.com/", "facebook.svg"},
			"?",
		},
		{
			"Twitter Secret Key",
			"Twitter Secret Key",
			"(?i)twitter(.{0,20})?[0-9a-z]{35,44}",
			[]string{},
			"?",
		},
		{
			"Twitter Client ID",
			"Twitter Client ID",
			"(?i)twitter(.{0,20})?[0-9a-z]{18,25}",
			[]string{},
			"?",
		},
		{
			"Github Personal Access Token",
			"Github Personal Access Token",
			"ghp_[0-9a-zA-Z]{36}",
			[]string{},
			"?",
		},
		{
			"Github OAuth Access Token",
			"Github OAuth Access Token",
			"gho_[0-9a-zA-Z]{36}",
			[]string{},
			"?",
		},
		{
			"Github App Token",
			"Github App Token",
			"(ghu|ghs)_[0-9a-zA-Z]{36}",
			[]string{},
			"?",
		},
		{
			"Github Refresh Token",
			"Github Refresh Token",
			"ghr_[0-9a-zA-Z]{76}",
			[]string{},
			"?",
		},
		{
			"LinkedIn Client ID",
			"LinkedIn Client ID",
			"(?i)linkedin(.{0,20})?(?-i)[0-9a-z]{12}",
			[]string{"linkedin.com/", "linkedin.svg"},
			"?",
		},
		{
			"LinkedIn Secret Key",
			"LinkedIn Secret Key",
			"(?i)linkedin(.{0,20})?[0-9a-z]{16}",
			[]string{"linkedin.com/", "linkedin.svg"},
			"?",
		},
		{
			"Slack",
			"Slack",
			"xox[baprs]-([0-9a-zA-Z]{10,48})?",
			[]string{},
			"?",
		},
		{
			"Asymmetric Private Key",
			"Asymmetric Private Key",
			"-----BEGIN ((EC|PGP|DSA|RSA|OPENSSH) )?PRIVATE KEY( BLOCK)?-----",
			[]string{},
			"?",
		},
		{
			"Google API key",
			"Google API key",
			"AIza[0-9A-Za-z\\-_]{35}",
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
			"(?i)heroku(.{0,20})?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
			[]string{},
			"?",
		},
		{
			"MailChimp API key",
			"MailChimp API key",
			"(?i)(mailchimp|mc)(.{0,20})?[0-9a-f]{32}-us[0-9]{1,2}",
			[]string{},
			"?",
		},
		{
			"Mailgun API key",
			"Mailgun API key",
			"((?i)(mailgun|mg)(.{0,20})?)?key-[0-9a-z]{32}",
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
			"sk_live_[0-9a-z]{32}",
			[]string{},
			"?",
		},
		{
			"SendGrid API Key",
			"SendGrid API Key",
			`SG\.[\w_]{16,32}\.[\w_]{16,64}`,
			[]string{},
			"?",
		},
		{
			"Slack Webhook",
			"Slack Webhook",
			"https://hooks.slack.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8,12}/[a-zA-Z0-9_]{24}",
			[]string{},
			"?",
		},
		{
			"Stripe API key",
			"Stripe API key",
			"(?i)stripe(.{0,20})?[sr]k_live_[0-9a-zA-Z]{24}",
			[]string{},
			"?",
		},
		{
			"Square access token",
			"Square access token",
			`sq0atp-[0-9A-Za-z\-_]{22}`,
			[]string{},
			"?",
		},
		{
			"Square OAuth secret",
			"Square OAuth secret",
			"sq0csp-[0-9A-Za-z\\-_]{43}",
			[]string{},
			"?",
		},
		{
			"Twilio API key",
			"Twilio API key",
			"(?i)twilio(.{0,20})?SK[0-9a-f]{32}",
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
			"shpss_[a-fA-F0-9]{32}",
			[]string{},
			"?",
		},
		{
			"Shopify access token",
			"Shopify access token",
			"shpat_[a-fA-F0-9]{32}",
			[]string{},
			"?",
		},
		{
			"Shopify custom app access token",
			"Shopify custom app access token",
			"shpca_[a-fA-F0-9]{32}",
			[]string{},
			"?",
		},
		{
			"Shopify private app access token",
			"Shopify private app access token",
			"shppa_[a-fA-F0-9]{32}",
			[]string{},
			"?",
		},
		{
			"PyPI upload token",
			"PyPI upload token",
			"pypi-AgEIcHlwaS5vcmc[A-Za-z0-9-_]{50,1000}",
			[]string{},
			"?",
		},
		{
			"S3 Bucket",
			"S3 Bucket",
			`(?:[a-zA-Z0-9_-]+s3.amazonaws.com|[a-zA-Z0-9_.-]+amazonaws.com|` +
				`[a-zA-Z0-9-\.\_]+\.s3\.amazonaws\.com|s3://[a-zA-Z0-9-\.\_]+|` +
				`s3-[a-zA-Z0-9-\.\_\/]+|s3.amazonaws.com/[a-zA-Z0-9-\.\_]+)`,
			[]string{},
			"?",
		},
	}
	return regexes
}

//RemoveDuplicateSecrets removes duplicates from secrets found
func RemoveDuplicateSecrets(input []SecretMatched) []SecretMatched {
	keys := make(map[string]bool)
	list := []SecretMatched{}
	for _, entry := range input {
		if _, value := keys[entry.Url]; !value {
			keys[entry.Url] = true
			list = append(list, entry)
		}
	}
	return list
}
