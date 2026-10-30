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

package output

import (
	"bytes"
	_ "embed"
	"html/template"
	"os"
	"strings"

	fileUtils "github.com/edoardottt/cariddi/v2/internal/file"
	"github.com/edoardottt/cariddi/v2/pkg/scanner"
)

// The report is a single file: style and script are inlined by the template.
var (
	//go:embed templates/report.html.tmpl
	reportTemplate string
	//go:embed templates/report.css
	reportStyle string
	//go:embed templates/report.js
	reportScript string
)

// htmlReport is the data rendered by templates/report.html.tmpl.
// Empty sections (disabled flag or nothing found) are not shown,
// except for the results.
type htmlReport struct {
	Generated  string
	Stats      []htmlStat
	Results    []string
	Secrets    []scanner.SecretMatched
	Endpoints  []scanner.EndpointMatched
	Extensions []scanner.FileTypeMatched
	Errors     []scanner.ErrorMatched
	Infos      []scanner.InfoMatched
}

// htmlStat is a summary card.
type htmlStat struct {
	ID    string
	Label string
	Count int
}

// htmlSection is the header of a results table.
// ID is used both as anchor and as style hook (see the k-* CSS classes).
type htmlSection struct {
	ID      string
	Title   string
	Count   int
	Columns []string
}

// newHTMLSection returns the header of a results table.
func newHTMLSection(id, title string, count int, columns ...string) htmlSection {
	return htmlSection{ID: id, Title: title, Count: count, Columns: columns}
}

// isHTTP reports whether url is an http(s) URL.
func isHTTP(url string) bool {
	lower := strings.ToLower(url)

	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

// writeHTML renders the report into the html file.
// html/template escapes the findings according to the context they are in.
func writeHTML(filename string, report htmlReport) error {
	tmpl, err := template.New("report.html.tmpl").Funcs(template.FuncMap{
		"style":    func() template.CSS { return template.CSS(reportStyle) },
		"script":   func() template.JS { return template.JS(reportScript) },
		"logoMark": func() string { return logoMarkPath },
		"logoText": func() string { return logoTextPath },
		"section":  newHTMLSection,
		"isHTTP":   isHTTP,
	}).Parse(reportTemplate)
	if err != nil {
		return err
	}

	// Render the whole report first, a failure must not leave a truncated file.
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, report); err != nil {
		return err
	}

	return os.WriteFile(filename, buf.Bytes(), fileUtils.Permission0644)
}
