package processor

import (
	"bytes"
	"sync"
)

// FileJob is a struct used to hold all of the results of processing internally before sent to the formatter
type FileJob struct {
	Filename  string
	Extension string
	Location  string
	Content   []byte
	Bytes     int64
	Hash      []byte
	Binary    bool
	Score     float64
	Locations map[string][]int
}

// CheckDuplicates is used to hold hashes if duplicate detection is enabled it comes with a mutex
// that should be locked while a check is being performed then added
type CheckDuplicates struct {
	hashes map[int64][][]byte
	mux    sync.Mutex
}

// Non thread safe add a key into the duplicates check need to use mutex inside struct before calling this
func (c *CheckDuplicates) Add(key int64, hash []byte) {
	hashes, ok := c.hashes[key]
	if ok {
		c.hashes[key] = append(hashes, hash)
	} else {
		c.hashes[key] = [][]byte{hash}
	}
}

// Non thread safe check to see if the key exists already need to use mutex inside struct before calling this
func (c *CheckDuplicates) Check(key int64, hash []byte) bool {
	hashes, ok := c.hashes[key]
	if ok {
		for _, h := range hashes {
			if bytes.Equal(h, hash) {
				return true
			}
		}
	}

	return false
}
