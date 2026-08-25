// SPDX-License-Identifier: MIT

package main

import (
	"testing"
)

// The cache is keyed by root as well as query. Without that, a search of one
// directory would serve its file list to an identical query against another.
func TestSearchCacheIsolatesRoots(t *testing.T) {
	cache := NewSearchCache()
	cache.Store("/repo/a", nil, "token", []string{"/repo/a/one.go"})

	if _, ok := cache.FindPrefixFiles("/repo/b", nil, "token"); ok {
		t.Fatal("root /repo/b must not hit an entry stored for /repo/a")
	}

	files, ok := cache.FindPrefixFiles("/repo/a", nil, "token")
	if !ok {
		t.Fatal("expected a hit for the root the entry was stored under")
	}
	if len(files) != 1 || files[0] != "/repo/a/one.go" {
		t.Errorf("unexpected cached files: %v", files)
	}
}

// Prefix matching must still work within a single root, and must not be fooled
// by a root containing spaces — the query walk splits on whitespace.
func TestSearchCachePrefixMatchWithinRoot(t *testing.T) {
	cache := NewSearchCache()
	root := "/home/me/my projects/repo"
	cache.Store(root, nil, "alpha", []string{"/x/one.go"})

	files, ok := cache.FindPrefixFiles(root, nil, "alpha beta")
	if !ok {
		t.Fatal("expected the shorter prefix 'alpha' to hit")
	}
	if len(files) != 1 || files[0] != "/x/one.go" {
		t.Errorf("unexpected cached files: %v", files)
	}

	if _, ok := cache.FindPrefixFiles("/home/me/my", nil, "projects/repo alpha"); ok {
		t.Fatal("a root with spaces must not be confusable with another root plus query words")
	}
}

// Extension filters remain part of the key alongside the root.
func TestSearchCacheIsolatesExtensions(t *testing.T) {
	cache := NewSearchCache()
	cache.Store("/repo", []string{"go"}, "token", []string{"/repo/one.go"})

	if _, ok := cache.FindPrefixFiles("/repo", []string{"ts"}, "token"); ok {
		t.Fatal("a different extension filter must not share cache entries")
	}
	if _, ok := cache.FindPrefixFiles("/repo", []string{"go"}, "token"); !ok {
		t.Fatal("expected a hit for the same root and extensions")
	}
}
