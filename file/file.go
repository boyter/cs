// Package file provides file operations specific to code repositories
// such as walking the file tree obeying .ignore and .gitignore files
// or looking for the root directory assuming already in a git project

// SPDX-License-Identifier: MIT OR Unlicense

package file

import (
	"errors"
	"github.com/boyter/cs/processor/go-gitignore"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var TerminateWalkError = errors.New("walker terminated")

type File struct {
	Location string
	Filename string
}

type FileWalker struct {
	walkMutex              sync.Mutex
	terminateWalking       bool
	isWalking              bool
	directory              string
	fileListQueue          chan *File
	LocationExcludePattern []string // Case sensitive patterns which exclude files
	PathExclude            []string // Paths to always ignore such as .git,.svn and .hg
	IgnoreIgnoreFile       bool     // Should .ignore files be respected?
	IgnoreGitIgnore        bool     // Should .gitignore files be respected?
	IncludeHidden          bool     // Should hidden files and directories be included/walked
	InstanceId             int
	AllowListExtensions    []string // Which extensions should be allowed
	UniqueId               string
}

func NewFileWalker(directory string, fileListQueue chan *File) *FileWalker {
	return &FileWalker{
		walkMutex:              sync.Mutex{},
		fileListQueue:          fileListQueue,
		directory:              directory,
		terminateWalking:       false,
		isWalking:              false,
		LocationExcludePattern: []string{},
		PathExclude:            []string{},
		IgnoreIgnoreFile:       false,
		IncludeHidden:          false,
		AllowListExtensions:    []string{},
	}
}

// Call to get the state of the file walker and determine
// if we are walking or not
func (f *FileWalker) Walking() bool {
	f.walkMutex.Lock()
	defer f.walkMutex.Unlock()
	return f.isWalking
}

// Call to have the walker break out of walking and return as
// soon as it possibly can. This is needed because
// this walker needs to work in a TUI interactive mode and
// as such we need to be able to end old processes
func (f *FileWalker) Terminate() {
	f.walkMutex.Lock()
	defer f.walkMutex.Unlock()
	f.terminateWalking = true
}

// Starts walking the supplied directory with the supplied settings
// and putting files that mach into the supplied slice
// Returns usual ioutil errors if there is a file issue
// and a TerminateWalkError if terminate is called while walking
func (f *FileWalker) Start() error {
	f.walkMutex.Lock()
	f.isWalking = true
	f.walkMutex.Unlock()

	err := f.walkDirectoryRecursive(f.directory, []gitignore.IgnoreMatcher{})
	close(f.fileListQueue)

	f.walkMutex.Lock()
	f.isWalking = false
	f.walkMutex.Unlock()

	return err
}

func (f *FileWalker) walkDirectoryRecursive(directory string, ignores []gitignore.IgnoreMatcher) error {
	// NB have to call unlock not using defer because method is recursive
	// and will deadlock if not done manually
	f.walkMutex.Lock()
	if f.terminateWalking == true {
		f.walkMutex.Unlock()
		return TerminateWalkError
	}
	f.walkMutex.Unlock()

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
	// and any subdirectories
	for _, file := range files {
		if !f.IgnoreGitIgnore {
			if file.Name() == ".gitignore" {
				ignore, err := gitignore.NewGitIgnore(filepath.Join(directory, file.Name()))
				if err == nil {
					ignores = append(ignores, ignore)
				}
			}
		}

		if !f.IgnoreIgnoreFile {
			if file.Name() == ".ignore" {
				ignore, err := gitignore.NewGitIgnore(filepath.Join(directory, file.Name()))
				if err == nil {
					ignores = append(ignores, ignore)
				}
			}
		}
	}

	// Process files first to start feeding whatever process is consuming
	// the output before traversing into directories for more files
	for _, file := range files {
		shouldIgnore := false

		// Check against the ignore files we have if the file we are looking at
		// should be ignored
		// It is safe to always call this because the ignores will not be added
		// in previous steps
		for _, ignore := range ignores {
			if ignore.Match(filepath.Join(directory, file.Name()), file.IsDir()) {
				shouldIgnore = true
			}
		}

		// Ignore hidden files
		if !f.IncludeHidden {
			s, err := IsHidden(file, directory)
			if s {
				shouldIgnore = true
			}
			if err != nil {
				return err
			}
		}

		// Check against extensions
		if len(f.AllowListExtensions) != 0 {
			ext := GetExtension(file.Name())

			a := false
			for _, v := range f.AllowListExtensions {
				if v == ext {
					a = true
				}
			}

			// try again because we could have one of those pesky ones such as something.spec.tsx
			ext = GetExtension(ext)
			for _, v := range f.AllowListExtensions {
				if v == ext {
					a = true
				}
			}

			if !a {
				shouldIgnore = true
			}
		}

		if !shouldIgnore {
			for _, p := range f.LocationExcludePattern {
				if strings.Contains(filepath.Join(directory, file.Name()), p) {
					shouldIgnore = true
				}
			}

			if !shouldIgnore {
				f.fileListQueue <- &File{
					Location: filepath.Join(directory, file.Name()),
					Filename: file.Name(),
				}
			}
		}
	}

	// Now we process the directories after hopefully giving the
	// channel some files to process
	for _, dir := range dirs {
		shouldIgnore := false

		// Check against the ignore files we have if the file we are looking at
		// should be ignored
		// It is safe to always call this because the ignores will not be added
		// in previous steps
		for _, ignore := range ignores {
			if ignore.Match(filepath.Join(directory, dir.Name()), dir.IsDir()) {
				shouldIgnore = true
			}
		}

		// Confirm if there are any files in the path deny list which usually includes
		// things like .git .hg and .svn
		for _, deny := range f.PathExclude {
			if strings.HasSuffix(dir.Name(), deny) {
				shouldIgnore = true
			}
		}

		// Ignore hidden directories
		if !f.IncludeHidden {
			s, err := IsHidden(dir, directory)
			if s {
				shouldIgnore = true
			}
			if err != nil {
				return err
			}
		}

		if !shouldIgnore {
			for _, p := range f.LocationExcludePattern {
				if strings.Contains(filepath.Join(directory, dir.Name()), p) {
					shouldIgnore = true
				}
			}

			err = f.walkDirectoryRecursive(filepath.Join(directory, dir.Name()), ignores)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// Walk the supplied directory backwards looking for .git or .hg
// directories indicating we should start our search from that
// location as its the root.
// Returns the first directory below supplied with .git or .hg in it
// otherwise the supplied directory
func FindRepositoryRoot(startDirectory string) string {
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
	// rather than the root
	return startDirectory
}

// Check if there is a .git or .hg folder in the supplied directory
func checkForGitOrMercurial(curdir string) bool {
	if stat, err := os.Stat(filepath.Join(curdir, ".git")); err == nil && stat.IsDir() {
		return true
	}

	if stat, err := os.Stat(filepath.Join(curdir, ".hg")); err == nil && stat.IsDir() {
		return true
	}

	return false
}

// A custom version of extracting extensions for a file
// which deals with extensions specific to code such as
// .travis.yml and the like
func GetExtension(name string) string {

	name = strings.ToLower(name)
	if !strings.Contains(name, ".") {
		return name
	}

	if strings.LastIndex(name, ".") == 0 {
		return name
	}

	return path.Ext(name)[1:]
}
