package processor

import (
	"github.com/boyter/cs/file"
	"io"
	"io/ioutil"
	"os"
	//"sync"
	"sync/atomic"
)

type FileReaderWorker2 struct {
	input      chan *file.File
	output     chan *fileJob
	fileCount  int64 // Count of the number of files that have been read
	InstanceId int
}

func NewFileReaderWorker(input chan *file.File, output chan *fileJob) FileReaderWorker2 {
	return FileReaderWorker2{
		input:     input,
		output:    output,
		fileCount: 0,
	}
}

func (f *FileReaderWorker2) GetFileCount() int64 {
	return atomic.LoadInt64(&f.fileCount)
}

// This is responsible for spinning up all of the jobs
// that read files from disk into memory
func (f *FileReaderWorker2) Start() {

	for res := range f.input {
		fi, err := os.Stat(res.Location)
		if err != nil {
			continue
		}

		var content []byte
		var s int64 = 1024000

		// TODO we should NOT do this and instead use a scanner later on
		// Only read up to ~1MB of a file because anything beyond that is probably pointless
		if fi.Size() < s {
			content, err = ioutil.ReadFile(res.Location)
		} else {
			r, err := os.Open(res.Location)
			if err != nil {
				continue
			}

			var tmp [1024000]byte
			_, _ = io.ReadFull(r, tmp[:])
			_ = r.Close()
		}

		if err == nil {
			atomic.AddInt64(&f.fileCount, 1)
			f.output <- &fileJob{
				Filename:       res.Filename,
				Extension:      "",
				Location:       res.Location,
				Content:        content,
				Bytes:          0,
				Score:          0,
				MatchLocations: map[string][][]int{},
			}
		}
	}
			
	close(f.output)
}
