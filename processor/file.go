// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense
package processor

import (
	sccprocessor "github.com/boyter/scc/processor"
	"github.com/monochromegane/go-gitignore"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var WalkMutex = sync.Mutex{}          // We only ever want 1 file walker operating
var IsWalking = NewBool(false)        // The state indicating if we are walking
var TerminateWalking = NewBool(false) // The flag to indicate we should stop

func walkDirectory(directory string, fileListQueue chan *fileJob) error {
	WalkMutex.Lock()
	defer WalkMutex.Unlock()

	IsWalking.SetTo(true)
	err := walkDirectoryRecursive(directory, []gitignore.IgnoreMatcher{}, fileListQueue)
	close(fileListQueue)
	TerminateWalking.SetTo(false)
	IsWalking.SetTo(false)
	return err
}

// Walks a directory recursively using gitignore/ignore files to ignore files and directories
// as well as using extension checks to ensure only files that should be processed are
// let though.
func walkDirectoryRecursive(directory string, ignores []gitignore.IgnoreMatcher, fileListQueue chan *fileJob) error {
	// Because this can work in a interactive mode we need a way to be able
	// to stop walking such as when the user starts a new search which this return should
	// take care of
	if TerminateWalking.IsSet() == true {
		return nil
	}

	fileInfos, err := ioutil.ReadDir(directory)

	if err != nil {
		return err
	}

	files := []os.FileInfo{}
	dirs := []os.FileInfo{}

	// We want to break apart the files and directories from the
	// return as we loop over them differently and this avoids some
	// nested if logic at the expense of a "redundant" loop
	for _, file := range fileInfos {
		if file.IsDir() {
			dirs = append(dirs, file)
		} else {
			files = append(files, file)
		}
	}

	// Pull out all of the ignore and gitignore files and add them
	// to out collection of ignores to be applied for this pass
	// and later on
	if Ignore == false {
		for _, file := range files {
			if file.Name() == ".gitignore" || file.Name() == ".ignore" {
				ignore, err := gitignore.NewGitIgnore(filepath.Join(directory, file.Name()))
				if err == nil {
					ignores = append(ignores, ignore)
				}
			}
		}
	}

	for _, file := range files {
		shouldIgnore := false
		for _, ignore := range ignores {
			if ignore.Match(filepath.Join(directory, file.Name()), file.IsDir()) {
				shouldIgnore = true
			}
		}

		if !shouldIgnore {
			language, ext := sccprocessor.DetectLanguage(file.Name())

			// At this point we have passed all the ignore file checks
			// so now we are checking if there are extensions we should
			// be looking for
			if len(AllowListExtensions) != 0 {
				shouldIgnore = true
				for _, e := range AllowListExtensions {
					if ext == e {
						shouldIgnore = false
					}
				}
			}

			for _, p := range LocationExcludePattern {
				if strings.Contains(filepath.Join(directory, file.Name()), p) {
					shouldIgnore = true
				}
			}

			// We need to check the #! because any file without an extension is
			// considered a possible #! file
			// TODO we should allow those though and handle it later on
			if !shouldIgnore && len(language) != 0 && language[0] != "#!" {
				fileListQueue <- &fileJob{
					Location:  filepath.Join(directory, file.Name()),
					Filename:  file.Name(),
					Extension: ext,
					Locations: map[string][]int{},
				}
			}
		}
	}

	for _, dir := range dirs {
		shouldIgnore := false
		for _, ignore := range ignores {
			if ignore.Match(filepath.Join(directory, dir.Name()), dir.IsDir()) {
				shouldIgnore = true
			}
		}

		for _, deny := range PathDenylist {
			if strings.HasSuffix(dir.Name(), deny) {
				shouldIgnore = true
			}
		}

		if !shouldIgnore {
			err = walkDirectoryRecursive(filepath.Join(directory, dir.Name()), ignores, fileListQueue)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
