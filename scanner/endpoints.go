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

//EndpointMatched struct
type EndpointMatched struct {
	Parameters []Parameter
	Url        string
}

type Parameter struct {
	Parameter string
	Attacks   []string
}

//GetJuicyParameters returns juicy parameters
func GetJuicyParameters() []Parameter {
	var juicyParameters = []Parameter{
		{"apikey", []string{"Info"}},
		{"api_key", []string{"Info"}},
		{"api-key", []string{"Info"}},
		{"key", []string{"Info", "XSS"}},
		{"token", []string{"Info"}},
		{"secret", []string{"Info"}},
		{"user-id", []string{"Info"}},
		{"user_id", []string{"Info"}},
		{"userid", []string{"Info"}},
		{"auth", []string{"Info"}},
		{"admin", []string{"Info"}},
		{"dashboard", []string{"Info"}},
		{"manage", []string{"Info"}},
		{"debug", []string{"Info"}},
		{"dbg", []string{"Info"}},
		{"uid", []string{"Info"}},
		{"root", []string{"Info"}},
		{"shell", []string{"Info"}},
		{"id", []string{"SQLi", "XSS"}},
		{"page", []string{"SQLi", "LFI", "SSRF", "XSS"}},
		{"dir", []string{"SQLi", "LFI", "SSRF"}},
		{"search", []string{"SQLi", "XSS"}},
		{"category", []string{"SQLi", "XSS"}},
		{"file", []string{"SQLi", "LFI"}},
		{"class", []string{"SQLi"}},
		{"url", []string{"SQLi", "OpenRedir", "SSRF", "XSS"}},
		{"news", []string{"SQLi"}},
		{"item", []string{"SQLi"}},
		{"menu", []string{"SQLi"}},
		{"lang", []string{"SQLi", "XSS"}},
		{"name", []string{"SQLi", "XSS"}},
		{"ref", []string{"SQLi"}},
		{"title", []string{"SQLi"}},
		{"view", []string{"SQLi", "LFI", "OpenRedir", "SSRF", "XSS"}},
		{"topic", []string{"SQLi"}},
		{"thread", []string{"SQLi"}},
		{"type", []string{"SQLi", "LFI", "XSS"}},
		{"date", []string{"SQLi", "LFI", "XSS"}},
		{"form", []string{"SQLi"}},
		{"join", []string{"SQLi"}},
		{"main", []string{"SQLi"}},
		{"nav", []string{"SQLi"}},
		{"region", []string{"SQLi"}},
		{"cat", []string{"LFI"}},
		{"action", []string{"LFI"}},
		{"board", []string{"LFI"}},
		{"detail", []string{"LFI"}},
		{"download", []string{"LFI"}},
		{"path", []string{"LFI", "SSRF"}},
		{"folder", []string{"LFI"}},
		{"prefix", []string{"LFI"}},
		{"include", []string{"LFI"}},
		{"inc", []string{"LFI"}},
		{"locate", []string{"LFI"}},
		{"show", []string{"LFI"}},
		{"doc", []string{"LFI"}},
		{"site", []string{"LFI", "SSRF"}},
		{"content", []string{"LFI"}},
		{"document", []string{"LFI"}},
		{"layout", []string{"LFI"}},
		{"mod", []string{"LFI"}},
		{"conf", []string{"LFI"}},
		{"next", []string{"OpenRedir", "SSRF"}},
		{"target", []string{"OpenRedir"}},
		{"rurl", []string{"OpenRedir", "SSRF"}},
		{"dest", []string{"OpenRedir", "SSRF"}},
		{"destination", []string{"OpenRedir"}},
		{"redir", []string{"OpenRedir", "SSRF"}},
		{"redirect_uri", []string{"OpenRedir", "SSRF"}},
		{"redirect_url", []string{"OpenRedir", "SSRF"}},
		{"redirect", []string{"OpenRedir", "SSRF"}},
		{"go", []string{"OpenRedir", "SSRF"}},
		{"return", []string{"OpenRedir", "SSRF"}},
		{"continue", []string{"OpenRedir", "SSRF"}},
		{"image_url", []string{"OpenRedir", "SSRF"}},
		{"returnTo", []string{"OpenRedir"}},
		{"return_to", []string{"OpenRedir"}},
		{"checkout_url", []string{"OpenRedir", "SSRF"}},
		{"return_path", []string{"OpenRedir"}},
		{"out", []string{"OpenRedir", "SSRF"}},
		{"exec", []string{"RCE"}},
		{"cmd", []string{"RCE"}},
		{"command", []string{"RCE"}},
		{"execute", []string{"RCE"}},
		{"ping", []string{"RCE"}},
		{"query", []string{"RCE", "XSS"}},
		{"reg", []string{"RCE"}},
		{"do", []string{"RCE"}},
		{"func", []string{"RCE"}},
		{"arg", []string{"RCE"}},
		{"jump", []string{"RCE"}},
		{"code", []string{"RCE"}},
		{"option", []string{"RCE"}},
		{"load", []string{"RCE"}},
		{"option", []string{"RCE"}},
		{"process", []string{"RCE"}},
		{"step", []string{"RCE"}},
		{"read", []string{"RCE", "LFI"}},
		{"function", []string{"RCE"}},
		{"req", []string{"RCE"}},
		{"feature", []string{"RCE"}},
		{"exe", []string{"RCE"}},
		{"module", []string{"RCE"}},
		{"payload", []string{"RCE"}},
		{"run", []string{"RCE"}},
		{"print", []string{"RCE"}},
		{"uri", []string{"SSRF"}},
		{"window", []string{"SSRF"}},
		{"data", []string{"SSRF"}},
		{"reference", []string{"SSRF"}},
		{"html", []string{"SSRF"}},
		{"val", []string{"SSRF"}},
		{"validate", []string{"SSRF"}},
		{"domain", []string{"SSRF"}},
		{"callback", []string{"SSRF"}},
		{"feed", []string{"SSRF"}},
		{"host", []string{"SSRF"}},
		{"port", []string{"SSRF"}},
		{"to", []string{"SSRF"}},
		{"host", []string{"SSRF"}},
		{"host", []string{"SSRF"}},
		{"q", []string{"XSS"}},
		{"keyword", []string{"XSS"}},
		{"keywords", []string{"XSS"}},
		{"month", []string{"XSS"}},
		{"year", []string{"XSS"}},
		{"email", []string{"XSS", "SQLi"}},
		{"terms", []string{"XSS"}},
		{"term", []string{"XSS"}},
		{"begindate", []string{"XSS"}},
		{"enddate", []string{"XSS"}},
		{"categoryid", []string{"XSS", "SQLi"}},
		{"p", []string{"XSS"}},
		{"l", []string{"XSS"}},
		{"s", []string{"XSS"}},
		{"list_type", []string{"XSS"}},
	}
	return juicyParameters
}

//RemovDuplicateEndpoints removes duplicate endpoints found
func RemovDuplicateEndpoints(input []EndpointMatched) []EndpointMatched {
	keys := make(map[string]bool)
	list := []EndpointMatched{}
	for _, entry := range input {
		if _, value := keys[entry.Url]; !value {
			keys[entry.Url] = true
			list = append(list, entry)
		}
	}
	return list
}
