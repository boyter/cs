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

func walkDirectoryRecursive(directory string, ignores []gitignore.IgnoreMatcher, fileListQueue chan *FileJob) error {
	if TerminateWalking.IsSet() == true {
		return nil
	}

	fileInfos, err := ioutil.ReadDir(directory)

	if err != nil {
		return err
	}

	files := []os.FileInfo{}
	dirs := []os.FileInfo{}

	for _, file := range fileInfos {
		if file.IsDir() {
			dirs = append(dirs, file)
		} else {
			files = append(files, file)
		}
	}

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

			if len(language) != 0 && language[0] != "#!" {
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
