// SPDX-License-Identifier: MIT

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// discoverGitRepos finds git repositories to sync. If dir itself contains
// a .git directory it is returned as the sole repo. Otherwise immediate
// child directories that contain .git are returned. No deeper recursion.
func discoverGitRepos(dir string) []string {
	if info, err := os.Stat(filepath.Join(dir, ".git")); err == nil && info.IsDir() {
		return []string{dir}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var repos []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		child := filepath.Join(dir, e.Name())
		if info, err := os.Stat(filepath.Join(child, ".git")); err == nil && info.IsDir() {
			repos = append(repos, child)
		}
	}
	return repos
}

// gitPull runs "git pull" in the given directory and returns combined output and any error.
func gitPull(repoDir string) (string, error) {
	cmd := exec.Command("git", "-C", repoDir, "pull")
	out, err := cmd.CombinedOutput()
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

	repos := discoverGitRepos(dir)
	if len(repos) == 0 {
		fmt.Fprintf(os.Stderr, "[git-sync] no git repositories found in %s\n", dir)
		return func() {}
	}

	workers := cfg.GitSyncWorkers
	if workers < 1 {
		workers = 1
	}

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
