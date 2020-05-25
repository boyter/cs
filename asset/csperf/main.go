// SPDX-License-Identifier: MIT
// SPDX-License-Identifier: Unlicense

package main

import (
	"fmt"
	str "github.com/boyter/cs/string"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

// Simple test comparison between various search methods
func main() {
	//f, _ := os.Create("csperf.pprof")
	//pprof.StartCPUProfile(f)
	//defer pprof.StopCPUProfile()

	arg1 := os.Args[1]
	arg2 := os.Args[2]

	b, err := ioutil.ReadFile(arg2)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Println("File length", len(b))

	haystack := string(b)

	var start time.Time
	var elapsed time.Duration

	fmt.Println("\nFindAllIndex (regex)")
	r := regexp.MustCompile(regexp.QuoteMeta(arg1))
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := r.FindAllIndex(b, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	fmt.Println("\nIndexAll (custom)")
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := str.IndexAll(haystack, arg1, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	r = regexp.MustCompile(`(?i)` + regexp.QuoteMeta(arg1))
	fmt.Println("\nFindAllIndex (regex ignore case)")
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := r.FindAllIndex(b, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	fmt.Println("\nIndexAllIgnoreCaseUnicode (custom)")
	for i := 0; i < 3; i++ {
		start = time.Now()
		all := str.IndexAllIgnoreCaseUnicode(haystack, arg1, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}
}
