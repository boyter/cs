package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/dbaggerman/cuba"
	"github.com/monochromegane/go-gitignore"
)

// Used as quick lookup for files with the same name to avoid some processing
// needs to be sync.Map as it potentially could be called by many GoRoutines
var extensionCache sync.Map

// A custom version of extracting extensions for a file
// which also has a case insensitive cache in order to save
// some needless processing
func getExtension(name string) string {
	name = strings.ToLower(name)
	extension, ok := extensionCache.Load(name)

	if ok {
		return extension.(string)
	}

	ext := filepath.Ext(name)

	if ext == "" || strings.LastIndex(name, ".") == 0 {
		extension = name
	} else {
		// Handling multiple dots or multiple extensions only needs to delete the last extension
		// and then call filepath.Ext.
		// If there are multiple extensions, it is the value of subExt,
		// otherwise subExt is an empty string.
		subExt := filepath.Ext(strings.TrimSuffix(name, ext))
		extension = strings.TrimPrefix(subExt+ext, ".")
	}

	extensionCache.Store(name, extension)
	return extension.(string)
}

type DirectoryJob struct {
	root    string
	path    string
	ignores []gitignore.IgnoreMatcher
}

type DirectoryWalker struct {
	buffer   *cuba.Pool
	output   chan<- *FileJob
	excludes []*regexp.Regexp
}

func NewDirectoryWalker(output chan<- *FileJob) *DirectoryWalker {
	directoryWalker := &DirectoryWalker{
		output: output,
	}
	for _, exclude := range Exclude {
		directoryWalker.excludes = append(directoryWalker.excludes, regexp.MustCompile(exclude))
	}

	directoryWalker.buffer = cuba.New(directoryWalker.Readdir, cuba.NewStack())

	return directoryWalker
}

func (dw *DirectoryWalker) Walk(root string) error {
	root = filepath.Clean(root)

	fileInfo, err := os.Stat(root)
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		fileJob := newFileJob(root, root)
		if fileJob != nil {
			dw.output <- fileJob
		}

		return nil
	}

	dw.buffer.Push(
		&DirectoryJob{
			root:    root,
			path:    root,
			ignores: nil,
		},
	)

	return nil
}

func (dw *DirectoryWalker) Run() {
	routineWaitGroup.Add(1)
	defer routineWaitGroup.Done()

	dw.buffer.Finish()
	close(dw.output)
}

func (dw *DirectoryWalker) Readdir(handle *cuba.Handle) {
	job := handle.Item().(*DirectoryJob)

	ignores := job.ignores

	file, err := os.Open(job.path)
	if err != nil {
		printError(fmt.Sprintf("failed to open %s: %v", job.path, err))
		return
	}
	defer file.Close()

	dirents, err := file.Readdir(-1)
	if err != nil {
		printError(fmt.Sprintf("failed to read %s: %v", job.path, err))
		return
	}

	for _, dirent := range dirents {
		if returnEarly() {
			dw.buffer.Abort()
			close(dw.output)
			return
		}

		name := dirent.Name()

		if (!GitIgnore && name == ".gitignore") || (!Ignore && name == ".ignore") {
			path := filepath.Join(job.path, name)

			ignore, err := gitignore.NewGitIgnore(path)
			if err != nil {
				printError(fmt.Sprintf("failed to load gitignore %s: %v", job.path, err))
			}
			ignores = append(ignores, ignore)
		}
	}

DIRENTS:
	for _, dirent := range dirents {
		if returnEarly() {
			dw.buffer.Abort()
			close(dw.output)
			return
		}

		name := dirent.Name()
		path := filepath.Join(job.path, name)
		isDir := dirent.IsDir()

		for _, black := range PathBlacklist {
			if strings.HasPrefix(path, filepath.Join(job.root, black)) {
				if Verbose {
					printWarn(fmt.Sprintf("skipping directory due to being in blacklist: %s", path))
				}
				continue DIRENTS
			}
		}

		for _, exclude := range dw.excludes {
			if exclude.Match([]byte(name)) {
				if Verbose {
					printWarn("skipping directory due to match exclude: " + name)
				}
				continue DIRENTS
			}
		}

		for _, ignore := range ignores {
			if ignore.Match(path, isDir) {
				if Verbose {
					printWarn("skipping directory due to ignore: " + path)
				}
				continue DIRENTS
			}
		}

		if isDir {
			handle.Push(
				&DirectoryJob{
					root:    job.root,
					path:    path,
					ignores: ignores,
				},
			)
		} else {
			fileJob := newFileJob(path, name)
			if fileJob != nil {
				dw.output <- fileJob
			}
		}
	}
}

func newFileJob(path, name string) *FileJob {
	if returnEarly() {
		return nil
	}

	extension := getExtension(name)

	if len(WhiteListExtensions) != 0 {
		ok := false
		for _, x := range WhiteListExtensions {
			if x == extension {
				ok = true
			}
		}

		if !ok {
			if Verbose {
				printWarn(fmt.Sprintf("skipping file as not whitelisted: %s", name))
			}
			return nil
		}
	}

	return &FileJob{
		Location: path,
		Filename: name,
	}
}
