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

//Info struct
// Name = the name that identifies the information
// Regex = The regular expression to be matched
type Info struct {
	Name  string
	Regex []string
}

//InfoMatched struct
// Info = Info struct
// Url = url in which the information is found
// Match = the string matching the regex
type InfoMatched struct {
	Info  Info
	URL   string
	Match string
}

//GetInfoRegexes returns all the info structs
func GetInfoRegexes() []Info {
	var regexes = []Info{
		{
			"Email address",
			[]string{
				`(?i)([a-zA-Z0-9_.+-]+@[a-zA-Z0-9]+[a-zA-Z0-9-]*\.[a-zA-Z0-9-.]*[a-zA-Z0-9]{2,})`},
		},
		{
			"HTML comment",
			[]string{
				`(?i)(\<![\s]*--[\-!@#$%^&*:;ºª.,"'(){}\w\s\/\\[\]]*--[\s]*\>)`},
		},
		{
			"Internal IP address",
			[]string{
				`((172\.\d{1,3}\.\d{1,3}\.\d{1,3})|(192\.168\.\d{1,3}\.\d{1,3})|(10\.\d{1,3}\.\d{1,3}\.\d{1,3})|([fF][eE][89aAbBcCdDeEfF]::))`},
		},
		{
			"IPv4 address",
			[]string{
				`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)){3}`},
		},
		{
			"BTC address",
			[]string{
				`[13][a-km-zA-HJ-NP-Z1-9]{25,34}`},
		},
		/*
			HOW TO AVOID VERY VERY LONG BASE64 IMAGES ???
				{
					"Base64-encoded JSON",
					[]string{
						`ey(A|B)[A-Za-z0-9+\/]{20,}(={0,2})`},
				},
		*/
	}
	return regexes
}

//RemoveDuplicateInfos removes duplicates from Infos found
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
