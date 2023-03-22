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

package utils

import (
	"net/http"
	"reflect"
	"testing"
	"net/url"
	"github.com/gocolly/colly"
	"github.com/edoardottt/cariddi/pkg/output"
	"github.com/edoardottt/cariddi/pkg/scanner"
)

func TestHasProtocol(t *testing.T) {
	headers := http.Header{}
	headers.Set("Content-Type", "application/pdf")
	headers.Set("Content-Length", "128")
	req := colly.Request{}
	u, _ := url.Parse("http://test.com.pdf/id=5")
	req.Method = "GET"
	req.URL = u
	tests := []struct {
		name  string
		r *colly.Response
		secrets []scanner.SecretMatched
		filetype *scanner.FileType
		parameters []scanner.Parameter
		errors []scanner.ErrorMatched
		infos []scanner.InfoMatched
		want string
	}{
		{
			name:  "ok1",
			r: &colly.Response{
				StatusCode: 200,
				Body: []byte("abcd"),
				Ctx: nil,
				Request: &req,
				Headers: &headers,
			},
			secrets: []scanner.SecretMatched{
				scanner.SecretMatched{
					Secret: scanner.Secret{
						Name: "mysecret",
						Description: "My Secret",
						Regex: "random.*regex",
						FalsePositives: []string{},
						Poc: "POC",
					},
					URL: "http://test.com/id=5",
					Match: "it's a random day for my secret regex to be found",
				},
			},
			parameters: []scanner.Parameter{
				scanner.Parameter{
					Parameter: "id",
					Attacks: []string{},
				},
			},
			filetype: &scanner.FileType{
				Extension: "pdf",
				Severity: 7,
			},
			errors: []scanner.ErrorMatched{
				scanner.ErrorMatched{
					Error: scanner.Error{
						ErrorName: "MySQL error",
						Regex: []string{"MySQL.*error"},
					},
					URL: "http://test.com/id=5",
					Match: "it is a MySQL error happening",
				},
			},
			infos: []scanner.InfoMatched{
				scanner.InfoMatched{
					Info: scanner.Info{
						Name: "info1",
						Regex: []string{"my.*great"},
					},
					URL: "http://test.com/id=5",
					Match: "its my pleasure to inform you on this great day",
				},
			},
			want: `{"url":"http://test.com.pdf/id=5","method":"GET","status_code":200,"words":1,"lines":1,"content_type":"application/pdf","content_length":128,"matches":{"filetype":{"extension":"pdf","severity":7},"parameters":[{"name":"id","attacks":[]}],"errors":[{"name":"MySQL error","match":"it is a MySQL error happening"}],"infos":[{"name":"info1","match":"its my pleasure to inform you on this great day"}],"secrets":[{"name":"mysecret","match":"it's a random day for my secret regex to be found"}]}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := output.GetJSONString(tt.r, tt.secrets, tt.parameters, tt.filetype, tt.errors, tt.infos); !reflect.DeepEqual(string(got), tt.want) {
				t.Errorf("GetJSONString\n%v", string(got))
				t.Errorf("want\n%v", tt.want)
			}
		})
	}
}
