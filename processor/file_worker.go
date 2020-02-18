package processor

import (
	"github.com/boyter/cs/file"
	"io"
	"io/ioutil"
	"os"
	//"sync"
	"sync/atomic"
)

// This is responsible for spinning up all of the jobs
// that read files from disk into memory
// TODO make this spawn goroutines
func fileReaderWorker(input chan *file.File, output chan *fileJob) {
	//var wg sync.WaitGroup
	TotalCount = 0 // TODO make this somewhere that isnt a global
	//for i := 0; i < FileReadJobWorkers; i++ {
	//	wg.Add(1)
	//
	//	go func() {
	for res := range input {
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
			atomic.AddInt64(&TotalCount, 1)
			output <- &fileJob{
				Filename:  res.Filename,
				Extension: "",
				Location:  res.Location,
				Content:   content,
				Bytes:     0,
				Hash:      nil,
				Binary:    false,
				Score:     0,
				Locations: nil,
				Minified:  false,
			}
		}
	}
	//	wg.Done()
	//}()
	close(output)
}
//
//wg.Wait()
//close(output)
//}
