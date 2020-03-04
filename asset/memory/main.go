package main

import (
	"bytes"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	pageSize := 1024
	totalSize := pageSize*1_000_000
	// Assume 1kb per log entry
	// 1024 bytes to KB
	// 1000000 means about 1GB
	fmt.Println(totalSize, "bytes")
	mySlice := make([]byte, totalSize)

	// warm to ensure timing is fair
	for j := 0; j < 5; j++ {
		for i := 0; i < len(mySlice); i += pageSize {
			bytes.Index(mySlice[i:i+pageSize], []byte("some sort of thing"))
		}
	}

	for j := 0; j < 5; j++ {
		start := time.Now().UnixNano() / int64(time.Millisecond)
		for i := 0; i < len(mySlice); i += pageSize {
			bytes.Index(mySlice[i:i+pageSize], []byte("some sort of thing"))
		}
		fmt.Println("Single Threaded", time.Now().UnixNano()/int64(time.Millisecond)-start, "ms")
	}

	// https://boyter.org/posts/my-personal-complaints-about-golang/
	// var input = make(chan []byte, len(toProcess))

	for j := 0; j < 5; j++ {
		start := time.Now().UnixNano() / int64(time.Millisecond)
		cpuCount := runtime.NumCPU()
		chunk := totalSize / cpuCount
		var wg sync.WaitGroup
		for i := 0; i < cpuCount; i++ {
			wg.Add(1)

			// Spawn a routine to search this chunk
			go func(i int, j int) {
				for k := i; k < j; k += pageSize {
					bytes.Index(mySlice[k:k+pageSize], []byte("some sort of thing"))
				}
				wg.Done()
			}(i*chunk, i*chunk+chunk)
		}
		wg.Wait()
		fmt.Println("Multi Threaded", time.Now().UnixNano()/int64(time.Millisecond)-start, "ms")
	}
}
