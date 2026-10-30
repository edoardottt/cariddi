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

package output_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/edoardottt/cariddi/v2/pkg/input"
	"github.com/edoardottt/cariddi/v2/pkg/output"
	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

func TestHTMLOutput(t *testing.T) {
	// HTMLOutput creates the output folder in the working directory.
	t.Chdir(t.TempDir())

	secrets := []scanner.SecretMatched{
		{
			Secret: scanner.Secret{Name: "mysecret"},
			URL:    "javascript:alert(1)",
			Match:  `token="<script>alert('x')</script>&"`,
		},
	}
	endpoints := []scanner.EndpointMatched{
		{
			URL: "http://test.com?id=5&q=1",
			Parameters: []scanner.Parameter{
				{Parameter: "id", Attacks: []string{"SQLi", "IDOR"}},
				{Parameter: "q", Attacks: []string{}},
			},
		},
	}
	extensions := []scanner.FileTypeMatched{
		{Filetype: scanner.FileType{Extension: "pdf", Severity: 7}, URL: "http://test.com/a.pdf"},
		{Filetype: scanner.FileType{Extension: "sql", Severity: 1}, URL: "http://test.com/a.sql"},
	}
	all := input.Input{Secrets: true, Endpoints: true, Extensions: 7}

	tests := []struct {
		name    string
		flags   input.Input
		results []string
		want    []string
		notWant []string
	}{
		{
			name:  "test_no_results",
			flags: all,
			want: []string{
				`<div class="stat k-results empty">`,
				// An empty table must match tbody:empty.
				`<span class="section-title">Results</span><span class="count">0</span></summary>` + "\n" +
					`<div class="table-wrap"><table><thead><tr><th>URL</th></tr></thead><tbody></tbody>`,
				"</footer>\n<script>\n(() => {",
				"</body>\n</html>\n",
			},
		},
		{
			name:    "test_escaping",
			flags:   all,
			results: []string{`http://test.com?q="><script>alert(1)</script>`},
			want: []string{
				`<a class="stat k-results" href="#results">`,
				`<td class="url"><a href="http://test.com?q=%22%3e%3cscript%3ealert%281%29%3c/script%3e" ` +
					`target="_blank" rel="noopener noreferrer">http://test.com?q=&#34;&gt;&lt;script&gt;alert(1)&lt;/script&gt;</a>`,
				`<td>mysecret</td><td><code>token=&#34;&lt;script&gt;alert(&#39;x&#39;)&lt;/script&gt;&amp;&#34;</code></td>`,
				// Only http(s) URLs are clickable.
				`<td class="url">javascript:alert(1)</td>`,
			},
			notWant: []string{"<script>alert", `href="javascript:`},
		},
		{
			name:  "test_endpoints_extensions",
			flags: all,
			want: []string{
				`<span class="param"><code>id</code> <span class="tag">SQLi</span> <span class="tag">IDOR</span></span>` +
					`<span class="sep">; </span><span class="param"><code>q</code></span></td></tr>`,
				// Juiciest extensions first.
				`<td><code>sql</code></td><td><span class="level lvl-1" title="1 (juicy) to 7 (not juicy)">1</span></td>`,
				`a.sql</a></td></tr>` + "\n" + `<tr><td><code>pdf</code></td>`,
			},
		},
		{
			name:    "test_flags_disabled",
			results: []string{"http://test.com"},
			want:    []string{`<section class="section k-results" id="results">`},
			notWant: []string{`id="secrets"`, `id="endpoints"`, `id="extensions"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), "report.html")
			output.HTMLOutput(tt.flags, filename, tt.results, secrets, endpoints, extensions, nil, nil)

			got, err := os.ReadFile(filename)
			if err != nil {
				t.Fatal(err)
			}

			for _, want := range tt.want {
				if !strings.Contains(string(got), want) {
					t.Errorf("HTMLOutput: missing\n%v", want)
				}
			}

			for _, notWant := range tt.notWant {
				if strings.Contains(string(got), notWant) {
					t.Errorf("HTMLOutput: unexpected\n%v", notWant)
				}
			}
		})
	}
}
