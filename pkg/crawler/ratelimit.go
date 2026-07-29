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

package crawler

import (
	"sync"
	"time"
)

type requestRateLimiter struct {
	mutex    sync.Mutex
	interval time.Duration
	next     time.Time
}

func newRequestRateLimiter(requestsPerSecond int) *requestRateLimiter {
	rate := time.Duration(requestsPerSecond)

	interval := time.Second / rate
	if time.Second%rate != 0 {
		interval++
	}

	return &requestRateLimiter{
		interval: interval,
	}
}

// Wait spaces request starts across every concurrent collector worker.
// The first request proceeds immediately; later requests start no faster
// than the configured interval.
func (limiter *requestRateLimiter) Wait() {
	limiter.mutex.Lock()
	defer limiter.mutex.Unlock()

	now := time.Now()
	if wait := limiter.next.Sub(now); wait > 0 {
		time.Sleep(wait)

		now = time.Now()
	}

	if now.Before(limiter.next) {
		now = limiter.next
	}

	limiter.next = now.Add(limiter.interval)
}
