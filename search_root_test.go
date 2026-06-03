// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindRepositoryRoot(t *testing.T) {
	tmp := t.TempDir()

	// Parent repo with a real .git directory.
	parent := filepath.Join(tmp, "parent")
	if err := os.MkdirAll(filepath.Join(parent, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	// Nested worktree at parent/.agent-shell/wt with .git as a regular file
	// (as `git worktree add` produces).
	worktree := filepath.Join(parent, ".agent-shell", "wt")
	if err := os.MkdirAll(worktree, 0o755); err != nil {
		t.Fatal(err)
	}
	gitFile := filepath.Join(worktree, ".git")
	if err := os.WriteFile(gitFile, []byte("gitdir: "+parent+"/.git/worktrees/wt\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Subdirectory inside the worktree, so we exercise the upward walk.
	worktreeSub := filepath.Join(worktree, "sub", "dir")
	if err := os.MkdirAll(worktreeSub, 0o755); err != nil {
		t.Fatal(err)
	}

	// Plain directory with no repo marker anywhere up the tree (under tmp).
	orphan := filepath.Join(tmp, "orphan")
	if err := os.MkdirAll(orphan, 0o755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		start string
		want  string
	}{
		{"parent repo root", parent, parent},
		{"worktree root", worktree, worktree},
		{"nested dir inside worktree", worktreeSub, worktree},
		{"no marker falls back to start", orphan, orphan},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findRepositoryRoot(tt.start)
			// Resolve symlinks so /var vs /private/var on macOS doesn't trip the comparison.
			gotR, _ := filepath.EvalSymlinks(got)
			wantR, _ := filepath.EvalSymlinks(tt.want)
			if gotR != wantR {
				t.Errorf("findRepositoryRoot(%q) = %q, want %q", tt.start, got, tt.want)
			}
		})
	}
}
