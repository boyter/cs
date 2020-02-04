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

type FileWalker struct {
	WalkMutex        sync.Mutex
	IsWalking        *AtomicBool
	TerminateWalking *AtomicBool
	Directory        string
	FileListQueue    chan *FileJob
}

func NewFileWalker(directory string, fileListQueue chan *FileJob) FileWalker {
	return FileWalker{
		WalkMutex:        sync.Mutex{},
		IsWalking:        NewBool(false),
		TerminateWalking: NewBool(false),
		Directory:        directory,
		FileListQueue:    fileListQueue,
	}
}

func (f *FileWalker) WalkDirectory() error {
	f.WalkMutex.Lock()
	defer f.WalkMutex.Unlock()

	f.IsWalking.SetTo(true)
	err := f.WalkDirectoryRecursive(f.Directory, []gitignore.IgnoreMatcher{})

	f.TerminateWalking.SetTo(false)
	f.IsWalking.SetTo(false)

	return err
}

func (f *FileWalker) WalkDirectoryRecursive(directory string, ignores []gitignore.IgnoreMatcher) error {
	// Because this can work in a interactive mode we need a way to be able
	// to stop walking such as when the user starts a new search which this return should
	// take care of
	if f.TerminateWalking.IsSet() == true {
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
	for _, file := range files {
		if file.Name() == ".gitignore" || file.Name() == ".ignore" {
			ignore, err := gitignore.NewGitIgnore(filepath.Join(directory, file.Name()))
			if err == nil {
				ignores = append(ignores, ignore)
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
				f.FileListQueue <- &FileJob{
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
			err = f.WalkDirectoryRecursive(filepath.Join(directory, dir.Name()), ignores)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func walkDirectory(directory string, fileListQueue chan *FileJob) error {
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
func walkDirectoryRecursive(directory string, ignores []gitignore.IgnoreMatcher, fileListQueue chan *FileJob) error {
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
				fileListQueue <- &FileJob{
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

// Walk the directory backwards looking for .git or .hg
// directories indicating we should start our search from that
// location as its the root
func findRepositoryRoot(startDirectory string) string {
	// Firstly try to determine our real location
	curdir, err := os.Getwd()
	if err != nil {
		return startDirectory
	}

	// Check if we have .git or .hg where we are and if
	// so just return because we are already there
	if checkForGitOrMercurial(curdir) {
		return startDirectory
	}

	// We did not find something, so now we need to walk the file tree
	// backwards in a cross platform way and if we find
	// a match we return that
	lastIndex := strings.LastIndex(curdir, string(os.PathSeparator))
	for lastIndex != -1 {
		curdir = curdir[:lastIndex]

		if checkForGitOrMercurial(curdir) {
			return curdir
		}

		lastIndex = strings.LastIndex(curdir, string(os.PathSeparator))
	}

	// If we didn't find a good match return the supplied directory
	// so that we start the search from where we started at least
	return startDirectory
}

func checkForGitOrMercurial(curdir string) bool {
	if stat, err := os.Stat(filepath.Join(curdir, ".git")); err == nil && stat.IsDir() {
		return true
	}

	if stat, err := os.Stat(filepath.Join(curdir, ".hg")); err == nil && stat.IsDir() {
		return true
	}

	return false
}
