// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// gitPullTimeout is the maximum time a single git pull is allowed to run.
const gitPullTimeout = 2 * time.Minute

// discoverGitRepos finds git repositories to sync. If dir itself contains
// a .git entry (directory or file, to support worktrees) it is returned as
// the sole repo. Otherwise immediate child directories containing .git are
// returned. No deeper recursion.
func discoverGitRepos(dir string) ([]string, error) {
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		return []string{dir}, nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory %s: %w", dir, err)
	}

	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(dir, e.Name())
		if _, err := os.Stat(filepath.Join(child, ".git")); err == nil {
			repos = append(repos, child)
		}
	}
	return repos, nil
}

// gitPull runs "git pull" in the given directory with a timeout and
// non-interactive environment. Returns combined output and any error.
func gitPull(repoDir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), gitPullTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", repoDir, "pull")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	cmd.Stdin = nil
	out, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("timed out after %s", gitPullTimeout)
	}
	return string(out), err
}

// syncRepos pulls all repos using a worker pool of the given size.
func syncRepos(repos []string, workers int, logf func(string, ...any)) {
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(r string) {
			defer wg.Done()
			defer func() { <-sem }()

			logf("[git-sync] pulling %s", r)
			out, err := gitPull(r)
			if err != nil {
				logf("[git-sync] error pulling %s: %v\n%s", r, err, out)
			} else {
				logf("[git-sync] pulled %s: %s", r, out)
			}
		}(repo)
	}
	wg.Wait()
}

// startGitSync discovers git repos under cfg.Directory and periodically
// pulls them. It runs an immediate sync, then repeats on cfg.GitSyncInterval.
// The returned stop function cancels the background ticker.
func startGitSync(cfg *Config) func() {
	dir := cfg.Directory
	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "[git-sync] error: cannot determine working directory: %v\n", err)
			return func() {}
		}
	}

	repos, err := discoverGitRepos(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[git-sync] error: %v\n", err)
		return func() {}
	}
	if len(repos) == 0 {
		fmt.Fprintf(os.Stderr, "[git-sync] no git repositories found in %s\n", dir)
		return func() {}
	}

	workers := cfg.GitSyncWorkers

	logf := func(format string, args ...any) {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}

	fmt.Fprintf(os.Stderr, "[git-sync] found %d repo(s), syncing every %s with %d worker(s)\n",
		len(repos), cfg.GitSyncInterval, workers)

	// Run sync in background so it doesn't block the server/TUI from starting
	ticker := time.NewTicker(cfg.GitSyncInterval)
	done := make(chan struct{})

	go func() {
		// Immediate sync
		syncRepos(repos, workers, logf)

		for {
			select {
			case <-ticker.C:
				syncRepos(repos, workers, logf)
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func() {
		close(done)
	}
}
