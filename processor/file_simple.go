package processor

import (
	sccprocessor "github.com/boyter/scc/processor"
	"os"
	"path/filepath"
)

func walkDirectorySimple(fileListQueue chan *FileJob) {
	_ = filepath.Walk("./", func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// TODO change this to recursive method and deal with all the ignore files etc...

		// Should we ignore it due to ignore files?
		if !info.IsDir() {

			language, ext := sccprocessor.DetectLanguage(filepath.Base(root))

			if len(language) != 0 && language[0] != "#!" {
				fileListQueue <- &FileJob{
					Location:  root,
					Filename:  filepath.Base(root),
					Extension: ext,
					Locations: map[string][]int{},
				}
			}
		}

		return nil
	})

	close(fileListQueue)
}
