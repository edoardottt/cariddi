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

package crawler

import (
	urlUtils "github.com/edoardottt/cariddi/internal/url"
	"github.com/edoardottt/cariddi/pkg/input"
	"github.com/edoardottt/cariddi/pkg/scanner"
)

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
	Cache            bool
	Debug            bool
	EndpointsFlag    bool
	ErrorsFlag       bool
	InfoFlag         bool
	Intensive        bool
	Rua              bool
	SecretsFlag      bool
	Ignore           string
	JSON             bool
	Proxy            string
	Target           string
	UserAgent        string
	FileType         int
	Headers          map[string]string
	StoreResp        bool
	MaxDepth         int
	IgnoreExtensions input.StringSlice

	// Settings
	Concurrency int
	Delay       int
	Timeout     int

	// Storage
	SecretsSlice   []string
	EndpointsSlice []string
	IgnoreTxtSlice []string
}

type Event struct {
	ProtocolTemp string
	TargetTemp   string
	Scope        *urlUtils.Scope
	Ignore       bool
	Debug        bool
	IgnoreSlice  []string
	// MaxDepth is the maximum depth to crawl (0 means no limit).
	// With a max depth, the pages are crawled depth by depth (see New).
	MaxDepth int

	// depth is the depth of the pages being crawled, with a max depth.
	// It changes only between two depths, when no request is running.
	depth int
	// next are the links to crawl at the next depth, with a max depth.
	next nextLinks
}
