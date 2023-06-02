// SPDX-License-Identifier: MIT
// https://github.com/bep/debounce/tree/master

package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebounce(t *testing.T) {
	var (
		counter1 uint64
		counter2 uint64
	)

	f1 := func() {
		atomic.AddUint64(&counter1, 1)
	}

	f2 := func() {
		atomic.AddUint64(&counter2, 1)
	}

	f3 := func() {
		atomic.AddUint64(&counter2, 2)
	}

	debounced := NewDebouncer(100 * time.Millisecond)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f1)
		}

		time.Sleep(200 * time.Millisecond)
	}

	for i := 0; i < 4; i++ {
		for j := 0; j < 10; j++ {
			debounced(f2)
		}
		for j := 0; j < 10; j++ {
			debounced(f3)
		}

		time.Sleep(200 * time.Millisecond)
	}

	c1 := int(atomic.LoadUint64(&counter1))
	c2 := int(atomic.LoadUint64(&counter2))
	if c1 != 3 {
		t.Error("Expected count 3, was", c1)
	}
	if c2 != 8 {
		t.Error("Expected count 8, was", c2)
	}
}

func TestDebounceConcurrentAdd(t *testing.T) {
	var wg sync.WaitGroup

	var flag uint64

	debounced := NewDebouncer(100 * time.Millisecond)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			debounced(func() {
				atomic.CompareAndSwapUint64(&flag, 0, 1)
			})
		}()
	}
	wg.Wait()

	time.Sleep(500 * time.Millisecond)
	c := int(atomic.LoadUint64(&flag))
	if c != 1 {
		t.Error("Flag not set")
	}
}

// Issue #1
func TestDebounceDelayed(t *testing.T) {

	var (
		counter1 uint64
	)

	f1 := func() {
		atomic.AddUint64(&counter1, 1)
	}

	debounced := NewDebouncer(100 * time.Millisecond)

	time.Sleep(110 * time.Millisecond)

	debounced(f1)

	time.Sleep(200 * time.Millisecond)

	c1 := int(atomic.LoadUint64(&counter1))
	if c1 != 1 {
		t.Error("Expected count 1, was", c1)
	}

}

func ExampleNew() {
	var counter uint64

	f := func() {
		atomic.AddUint64(&counter, 1)
	}

	debounced := NewDebouncer(100 * time.Millisecond)

	for i := 0; i < 3; i++ {
		for j := 0; j < 10; j++ {
			debounced(f)
		}

		time.Sleep(200 * time.Millisecond)
	}

	c := int(atomic.LoadUint64(&counter))

	fmt.Println("Counter is", c)
	// Output: Counter is 3
}
