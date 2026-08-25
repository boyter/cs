// SPDX-License-Identifier: MIT

package main

import (
	"sort"
	"strings"
	"time"

	"github.com/boyter/simplecache"
)

// SearchCache caches file locations (paths) keyed by query string so that
// progressive refinements of a query can skip the full filesystem walk
// and re-evaluate only the files that matched a shorter prefix of the query.
type SearchCache struct {
	cache *simplecache.Cache[[]string]
}

// NewSearchCache creates a SearchCache with LRU eviction, 100 items max, 60s TTL.
func NewSearchCache() *SearchCache {
	maxItems := 100
	evictionPolicy := simplecache.LRU
	evictionSamples := 5
	maxAge := 60 * time.Second

	return &SearchCache{
		cache: simplecache.New[[]string](simplecache.Option{
			MaxItems:        &maxItems,
			EvictionPolicy:  &evictionPolicy,
			EvictionSamples: &evictionSamples,
			MaxAge:          &maxAge,
		}),
	}
}

// FindPrefixFiles tries progressively shorter prefixes of the query to find
// cached file locations from a previous search. Returns the cached locations
// and true on hit, or nil and false on miss. The root scopes the lookup so
// that the same query run against two different directories never shares
// entries.
func (sc *SearchCache) FindPrefixFiles(root string, extensions []string, query string) ([]string, bool) {
	prefix := cacheKeyPrefix(root, extensions)

	// Try the full query first, then progressively shorter prefixes.
	// We split on spaces and try removing the last word each time.
	words := strings.Fields(query)
	for i := len(words); i >= 1; i-- {
		key := prefix + strings.Join(words[:i], " ")
		if files, ok := sc.cache.Get(key); ok {
			return files, true
		}
	}

	return nil, false
}

// Store saves the matched file locations under the full query key, scoped to
// the root the search walked.
func (sc *SearchCache) Store(root string, extensions []string, query string, locations []string) {
	key := cacheKeyPrefix(root, extensions) + query
	_ = sc.cache.Set(key, locations)
}

// cacheKeyPrefix builds a deterministic prefix from the search root and the
// extension filter so that searches over different directories, or with
// different ext= params, don't share cache entries.
//
// The root belongs in the prefix rather than the query string because
// FindPrefixFiles truncates the query word by word — a root containing a
// space would be shredded by that walk and could collide across directories.
func cacheKeyPrefix(root string, extensions []string) string {
	var sb strings.Builder
	sb.WriteString(root)
	sb.WriteByte(0)
	if len(extensions) != 0 {
		sorted := make([]string, len(extensions))
		copy(sorted, extensions)
		sort.Strings(sorted)
		sb.WriteString(strings.Join(sorted, ","))
		sb.WriteByte(':')
	}
	return sb.String()
}
