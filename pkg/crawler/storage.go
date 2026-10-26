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
	"sync"

	"github.com/gocolly/colly/v2/storage"
)

// visitedStorage is the colly in-memory storage with an atomic visited
// check. colly checks if a URL was visited and then marks it in two
// separate steps: two goroutines finding the same link at the same time
// would both request it.
type visitedStorage struct {
	*storage.InMemoryStorage

	mu      sync.Mutex
	visited map[uint64]struct{}
}

func newVisitedStorage() *visitedStorage {
	return &visitedStorage{InMemoryStorage: &storage.InMemoryStorage{}}
}

// Init initializes the storage.
func (s *visitedStorage) Init() error {
	s.visited = make(map[uint64]struct{})

	return s.InMemoryStorage.Init()
}

// IsVisited checks if a request was visited and marks it as visited
// (colly marks it right after a false result anyway). Note that the
// collector HasVisited methods mark the request too.
func (s *visitedStorage) IsVisited(requestID uint64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.visited[requestID]; ok {
		return true, nil
	}

	s.visited[requestID] = struct{}{}

	return false, nil
}

// Visited marks a request as visited.
func (s *visitedStorage) Visited(requestID uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.visited[requestID] = struct{}{}

	return nil
}
