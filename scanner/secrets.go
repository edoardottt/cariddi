package scanner

//Secret
type Secret struct {
	Name        string
	Description string
	Regex       string
	Poc         string
}

var regexes = map[string]Secret{
	Secret{
		"AWS Access Key",
		"AWS Access Key",
		"(A3T[A-Z0-9]|AKIA|AGPA|AIDA|AROA|AIPA|ANPA|ANVA|ASIA)[A-Z0-9]{16}",
		"?",
	},
	Secret{
		"AWS Secret Key",
		"AWS Secret Key",
		"(?i)aws(.{0,20})?(?-i)['\"][0-9a-zA-Z\/+]{40}['\"]",
		"?",
	},
	Secret{
		"AWS MWS Key",
		"AWS MWS Key",
		"amzn\.mws\.[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
		"?",
	},
	Secret{
		"Facebook Secret Key",
		"Facebook Secret Key",
		"(?i)(facebook|fb)(.{0,20})?(?-i)['\"][0-9a-f]{32}['\"]",
		"?",
	},
	Secret{
		"Facebook Client ID",
		"Facebook Client ID",
		"(?i)(facebook|fb)(.{0,20})?['\"][0-9]{13,17}['\"]",
		"?",
	},
	Secret{
		"Twitter Secret Key",
		"Twitter Secret Key",
		"(?i)twitter(.{0,20})?[0-9a-z]{35,44}",
		"?",
	},
	Secret{
		"Twitter Client ID",
		"Twitter Client ID",
		"(?i)twitter(.{0,20})?[0-9a-z]{18,25}",
		"?",
	},
	Secret{
		"Github Personal Access Token",
		"Github Personal Access Token",
		"ghp_[0-9a-zA-Z]{36}",
		"?",
	},
	Secret{
		"Github OAuth Access Token",
		"Github OAuth Access Token",
		"gho_[0-9a-zA-Z]{36}",
		"?",
	},
    Secret{
		"Github App Token",
		"Github App Token",
		"(ghu|ghs)_[0-9a-zA-Z]{36}",
		"?",
	},
	Secret{
		"Github Refresh Token",
		"Github Refresh Token",
		"ghr_[0-9a-zA-Z]{76}",
		"?",
	},
	Secret{
		"LinkedIn Client ID",
		"LinkedIn Client ID",
		"(?i)linkedin(.{0,20})?(?-i)[0-9a-z]{12}",
		"?",
	},
    Secret{
		"LinkedIn Secret Key",
		"LinkedIn Secret Key",
		"(?i)linkedin(.{0,20})?[0-9a-z]{16}",
		"?",
	},
	Secret{
		"Slack",
		"Slack",
		"xox[baprs]-([0-9a-zA-Z]{10,48})?",
		"?",
	},
    Secret{
		"Asymmetric Private Key",
		"Asymmetric Private Key",
		"-----BEGIN ((EC|PGP|DSA|RSA|OPENSSH) )?PRIVATE KEY( BLOCK)?-----",
		"?",
	},
    Secret{
		"Google API key",
		"Google API key",
		"AIza[0-9A-Za-z\\-_]{35}",
		"?",
	},
	Secret{
		"Google (GCP) Service Account",
		"Slack",
		`"type": "service_account"`,
		"?",
	},
    Secret{
		"Heroku API key",
		"Heroku API key",
		"(?i)heroku(.{0,20})?[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}",
		"?",
	},
    Secret{
		"MailChimp API key",
		"MailChimp API key",
		"(?i)(mailchimp|mc)(.{0,20})?[0-9a-f]{32}-us[0-9]{1,2}",
		"?",
	},
	Secret{
		"Mailgun API key",
		"Mailgun API key",
		"((?i)(mailgun|mg)(.{0,20})?)?key-[0-9a-z]{32}",
		"?",
	},
    Secret{
		"PayPal Braintree access token",
		"PayPal Braintree access token",
		"access_token\$production\$[0-9a-z]{16}\$[0-9a-f]{32}",
		"?",
	},
    Secret{
		"Picatic API key",
		"Picatic API key",
		"sk_live_[0-9a-z]{32}",
		"?",
	},
	Secret{
		"SendGrid API Key",
		"SendGrid API Key",
		"SG\.[\w_]{16,32}\.[\w_]{16,64}",
		"?",
	},
}

/*
    description = "Slack Webhook"
    regex = '''https://hooks.slack.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8,12}/[a-zA-Z0-9_]{24}'''
    tags = ["key", "slack"]
[[rules]]
    description = "Stripe API key"
    regex = '''(?i)stripe(.{0,20})?[sr]k_live_[0-9a-zA-Z]{24}'''
    tags = ["key", "Stripe"]
[[rules]]
    description = "Square access token"
    regex = '''sq0atp-[0-9A-Za-z\-_]{22}'''
    tags = ["key", "square"]
[[rules]]
    description = "Square OAuth secret"
    regex = '''sq0csp-[0-9A-Za-z\\-_]{43}'''
    tags = ["key", "square"]
[[rules]]
    description = "Twilio API key"
    regex = '''(?i)twilio(.{0,20})?SK[0-9a-f]{32}'''
    tags = ["key", "twilio"]
[[rules]]
    description = "Dynatrace ttoken"
    regex = '''dt0[a-zA-Z]{1}[0-9]{2}\.[A-Z0-9]{24}\.[A-Z0-9]{64}'''
    tags = ["key", "Dynatrace"]
[[rules]]
    description = "Shopify shared secret"
    regex = '''shpss_[a-fA-F0-9]{32}'''
    tags = ["key", "Shopify"]
[[rules]]
    description = "Shopify access token"
    regex = '''shpat_[a-fA-F0-9]{32}'''
    tags = ["key", "Shopify"]
[[rules]]
    description = "Shopify custom app access token"
    regex = '''shpca_[a-fA-F0-9]{32}'''
    tags = ["key", "Shopify"]
[[rules]]
    description = "Shopify private app access token"
    regex = '''shppa_[a-fA-F0-9]{32}'''
    tags = ["key", "Shopify"]
[[rules]]
    description = "PyPI upload token"
    regex = '''pypi-AgEIcHlwaS5vcmc[A-Za-z0-9-_]{50,1000}'''
    tags = ["key", "pypi"]

*/
