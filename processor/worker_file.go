// SPDX-License-Identifier: MIT OR Unlicense

package processor

import (
	"github.com/boyter/cs/file"
	"runtime"
	"sync"

	"io/ioutil"
	"os"

	"sync/atomic"
)

type FileReaderWorker struct {
	input            chan *file.File
	output           chan *FileJob
	fileCount        int64 // Count of the number of files that have been read
	InstanceId       int
	MaxReadSizeBytes int64
}

func NewFileReaderWorker(input chan *file.File, output chan *FileJob) *FileReaderWorker {
	return &FileReaderWorker{
		input:            input,
		output:           output,
		fileCount:        0,
		MaxReadSizeBytes: 10000000, // sensible default of 10MB decimal
	}
}

func (f *FileReaderWorker) GetFileCount() int64 {
	return atomic.LoadInt64(&f.fileCount)
}

// This is responsible for spinning up all of the jobs
// that read files from disk into memory
func (f *FileReaderWorker) Start() {
	var wg sync.WaitGroup

	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go func() {
			for res := range f.input {
				f.process(res)
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(f.output)
}

func (f *FileReaderWorker) process(res *file.File) {
	fi, err := os.Stat(res.Location)
	if err != nil {
		return
	}

	var content []byte

	// Only read up to ~10MB of a file because anything beyond that is probably pointless
	if fi.Size() < f.MaxReadSizeBytes {
		content, err = ioutil.ReadFile(res.Location)
	} else {
		fi, err := os.Open(res.Location)
		if err != nil {
			return
		}
		defer fi.Close()

		byteSlice := make([]byte, f.MaxReadSizeBytes)
		_, err = fi.Read(byteSlice)
		if err != nil {
			return
		}

		content = byteSlice
	}

	if err == nil {
		atomic.AddInt64(&f.fileCount, 1)
		f.output <- &FileJob{
			Filename:       res.Filename,
			Extension:      "",
			Location:       res.Location,
			Content:        content,
			Bytes:          len(content),
			Score:          0,
			MatchLocations: map[string][][]int{},
		}
	}
}
