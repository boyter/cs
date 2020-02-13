package file

import (
	sccprocessor "github.com/boyter/scc/processor"
	"github.com/monochromegane/go-gitignore"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type File struct {
	Location  string
	Filename  string
	Extension string
}

type FileWalker struct {
	walkMutex              sync.Mutex
	IsWalking              bool
	TerminateWalking       bool
	Directory              string
	FileListQueue          chan *File
	AllowListExtensions    []string
	LocationExcludePattern []string
	PathDenylist           []string
}

func NewFileWalker(directory string, fileListQueue chan *File) FileWalker {
	return FileWalker{
		walkMutex:              sync.Mutex{},
		IsWalking:              false,
		TerminateWalking:       false,
		Directory:              directory,
		FileListQueue:          fileListQueue,
		AllowListExtensions:    []string{},
		LocationExcludePattern: []string{},
		PathDenylist:           []string{},
	}
}

func (f *FileWalker) WalkDirectory() error {
	f.walkMutex.Lock()
	f.IsWalking = true
	f.walkMutex.Unlock()

	err := f.WalkDirectoryRecursive(f.Directory, []gitignore.IgnoreMatcher{})

	f.walkMutex.Lock()
	f.TerminateWalking = false
	f.IsWalking = false
	f.walkMutex.Unlock()

	return err
}

func (f *FileWalker) WalkDirectoryRecursive(directory string, ignores []gitignore.IgnoreMatcher) error {
	// Because this can work in a interactive mode we need a way to be able
	// to stop walking such as when the user starts a new search which this return should
	// take care of
	f.walkMutex.Lock()
	if f.TerminateWalking == true {
		f.walkMutex.Unlock()
		return nil
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
			if len(f.AllowListExtensions) != 0 {
				shouldIgnore = true
				for _, e := range f.AllowListExtensions {
					if ext == e {
						shouldIgnore = false
					}
				}
			}

			for _, p := range f.LocationExcludePattern {
				if strings.Contains(filepath.Join(directory, file.Name()), p) {
					shouldIgnore = true
				}
			}

			// We need to check the #! because any file without an extension is
			// considered a possible #! file
			// TODO we should allow those though and handle it later on
			if !shouldIgnore && len(language) != 0 && language[0] != "#!" {
				f.FileListQueue <- &File{
					Location:  filepath.Join(directory, file.Name()),
					Filename:  file.Name(),
					Extension: ext,
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

		for _, deny := range f.PathDenylist {
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
