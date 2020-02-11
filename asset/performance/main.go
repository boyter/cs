package main

import (
	"fmt"
	str "github.com/boyter/cs/string"
	"io/ioutil"
	"os"
	"regexp"
	"time"
)

func main() {
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

	fmt.Println("IndexAll")
	for i := 0; i < 5; i++ {
		start = time.Now()
		all := str.IndexAll(haystack, arg1, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	r := regexp.MustCompile(`(?i)` + arg1)
	fmt.Println("FindAllIndex (regex ignore case)")
	for i := 0; i < 5; i++ {
		start = time.Now()
		all := r.FindAllIndex(b, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}

	fmt.Println("IndexAllIgnoreCase")
	for i := 0; i < 5; i++ {
		start = time.Now()
		all := str.IndexAllIgnoreCase(haystack, arg1, -1)
		elapsed = time.Since(start)
		fmt.Println("Scan took", elapsed, len(all))
	}
}
