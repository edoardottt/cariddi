/*
==========
Cariddi
==========

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
at your option) any later version.

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

package crawler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/edoardottt/cariddi/pkg/crawler"
)

func TestCreateCollyAppliesRateLimitAcrossWorkers(t *testing.T) {
	var (
		mutex        sync.Mutex
		requestTimes []time.Time
	)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		mutex.Lock()

		requestTimes = append(requestTimes, time.Now())
		mutex.Unlock()

		writer.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	const (
		rateLimit    = 10
		requestCount = 4
	)

	target := strings.TrimPrefix(server.URL, "http://")
	collector := crawler.CreateCollyWithRateLimit(0, rateLimit, requestCount, 10, 0,
		false, false, false, "", "", target)

	for index := range requestCount {
		if err := collector.Visit(fmt.Sprintf("%s/%d", server.URL, index)); err != nil {
			t.Fatalf("queue request %d: %v", index, err)
		}
	}

	collector.Wait()

	mutex.Lock()
	defer mutex.Unlock()

	if len(requestTimes) != requestCount {
		t.Fatalf("server received %d requests, want %d", len(requestTimes), requestCount)
	}

	sort.Slice(requestTimes, func(first, second int) bool {
		return requestTimes[first].Before(requestTimes[second])
	})

	minimumSpan := 240 * time.Millisecond
	if span := requestTimes[len(requestTimes)-1].Sub(requestTimes[0]); span < minimumSpan {
		t.Fatalf("requests spanned %v, want at least %v", span, minimumSpan)
	}
}
