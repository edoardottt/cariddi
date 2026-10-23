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

package main

import (
	"context"
	"fmt"
	"math"
	"testing"
	"time"
)

// takeMustReturn runs Take in a goroutine and fails the test if it does
// not return within the given timeout.
func takeMustReturn(t *testing.T, limiter interface{ Take() }, timeout time.Duration) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		limiter.Take()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(timeout):
		t.Fatalf("Take() did not return within %v", timeout)
	}
}

func TestNewRateLimiterUnlimitedWhenRpsIsZero(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	limiter := newRateLimiter(ctx, 0)

	// An unlimited limiter must report an unbounded limit.
	if got, want := limiter.GetLimit(), uint(math.MaxUint32); got != want {
		t.Errorf("GetLimit() = %d, want %d (unlimited)", got, want)
	}

	// Take() must return immediately: a zero-rps token bucket would
	// otherwise block forever and hang the scan.
	takeMustReturn(t, limiter, time.Second)
}

func TestNewRateLimiterFiniteWhenRpsIsPositive(t *testing.T) {
	for _, rps := range []uint{1, 2, 5, 10} {
		t.Run(fmt.Sprintf("rps=%d", rps), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			limiter := newRateLimiter(ctx, rps)

			if got := limiter.GetLimit(); got != rps {
				t.Errorf("GetLimit() = %d, want %d", got, rps)
			}

			// Every configured token must be consumable without waiting
			// for the bucket to refill.
			for i := uint(0); i < rps; i++ {
				takeMustReturn(t, limiter, time.Second)
			}
		})
	}
}

func TestNewRateLimiterBlocksWhenBucketIsEmpty(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// A single-token bucket.
	limiter := newRateLimiter(ctx, 1)

	// First Take consumes the only token.
	takeMustReturn(t, limiter, time.Second)

	// The bucket is now empty: the next Take must wait for the refill
	// (which happens after 1 second) and stay blocked for ~200ms.
	blocked := make(chan struct{})
	start := time.Now()
	go func() {
		limiter.Take()
		close(blocked)
	}()

	select {
	case <-blocked:
		t.Fatalf("Take() returned after %v on an empty bucket, want it to block", time.Since(start))
	case <-time.After(200 * time.Millisecond):
		// Expected: the token bucket is empty and throttles the request.
	}
}

func TestNewRateLimiterDoesNotThrottleWithZeroTokens(t *testing.T) {
	// Regression test: a plain ratelimit.New(ctx, 0, ...) bucket has zero
	// tokens, so Take() blocks forever. newRateLimiter must map 0 to an
	// unlimited limiter instead.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	limiter := newRateLimiter(ctx, 0)

	done := make(chan struct{})
	go func() {
		for i := 0; i < 100; i++ {
			limiter.Take()
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("newRateLimiter(0) deadlocked on Take()")
	}
}
