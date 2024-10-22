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

package crawler

import "github.com/edoardottt/cariddi/pkg/scanner"

type Results struct {
	URLs       []string
	Secrets    []scanner.SecretMatched
	Endpoints  []scanner.EndpointMatched
	Extensions []scanner.FileTypeMatched
	Errors     []scanner.ErrorMatched
	Infos      []scanner.InfoMatched
}

type Scan struct {
	// Flags
	Cache         bool
	Debug         bool
	EndpointsFlag bool
	ErrorsFlag    bool
	InfoFlag      bool
	Intensive     bool
	Plain         bool
	Rua           bool
	SecretsFlag   bool
	Ignore        string
	IgnoreTxt     string
	JSON          bool
	HTML          string
	Proxy         string
	Target        string
	Txt           string
	UserAgent     string
	FileType      int
	Headers       map[string]string
	StoreResp     bool
	OutputDir     string

	// Settings
	Concurrency int
	Delay       int
	Timeout     int

	// Storage
	SecretsSlice   []string
	EndpointsSlice []string
}

type Event struct {
	ProtocolTemp string
	TargetTemp   string
	Target       string
	Intensive    bool
	Ignore       bool
	Debug        bool
	JSON         bool
	IgnoreSlice  []string
	URLs         *[]string
}
