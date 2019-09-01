package cuba

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type Task func(*Handle)

type Pool struct {
	mutex      *sync.Mutex
	bucket     Bucket
	cond       *sync.Cond
	numWorkers int32
	maxWorkers int32
	closed     bool
	task       Task
	wg         *sync.WaitGroup
}

// Constructs a new Cuba thread pool.
//
// The worker callback will be called by multiple goroutines in parallel, so is
// expected to be thread safe.
//
// Bucket affects the order that items will be processed in. cuba.NewQueue()
// provides FIFO ordering, while cuba.NewStack() provides LIFO ordered work.
func New(task Task, bucket Bucket) *Pool {
	m := &sync.Mutex{}
	return &Pool{
		mutex:      m,
		bucket:     bucket,
		cond:       sync.NewCond(m),
		task:       task,
		maxWorkers: int32(runtime.NumCPU()),
		wg:         &sync.WaitGroup{},
	}
}

// Sets the maximum number of worker goroutines.
//
// Default: runtime.NumCPU() (i.e. the number of CPU cores available)
func (pool *Pool) SetMaxWorkers(n int32) {
	pool.maxWorkers = n
}

// Push an item into the worker pool. This will be scheduled to run on a worker
// immediately.
func (pool *Pool) Push(item interface{}) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	// The ideal might be to have a fixed pool of worker goroutines which all
	// close down when the work is done.
	// However, since the bucket can drain down to 0 and appear done before the
	// final worker queues more items it's a little complicated.
	// Having a floating pool means we can restart workers as we discover more
	// work to be done, which solves this problem at the cost of a little
	// inefficiency.
	if pool.numWorkers < pool.maxWorkers {
		pool.wg.Add(1)
		go pool.runWorker()
	}

	pool.bucket.Push(item)
	pool.cond.Signal()
}

// Push multiple items into the worker pool.
//
// Compared to Push() this only aquires the lock once, so may reduce lock
// contention.
func (pool *Pool) PushAll(items []interface{}) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for i := 0; i < len(items); i++ {
		if pool.numWorkers >= pool.maxWorkers {
			break
		}
		pool.wg.Add(1)
		go pool.runWorker()
	}

	pool.bucket.PushAll(items)
	pool.cond.Broadcast()
}

// Calling Finish() waits for all work to complete, and allows goroutines to shut
// down.
func (pool *Pool) Finish() {
	pool.mutex.Lock()

	pool.closed = true
	pool.cond.Broadcast()

	pool.mutex.Unlock()
	pool.wg.Wait()
}

func (pool *Pool) next() (interface{}, bool) {
	pool.mutex.Lock()
	defer pool.mutex.Unlock()

	for pool.bucket.Empty() {
		if pool.closed {
			return nil, false
		}
		pool.cond.Wait()
	}

	item := pool.bucket.Pop()

	return item, true
}

func (pool *Pool) runWorker() {
	atomic.AddInt32(&pool.numWorkers, 1)
	handle := Handle{
		pool: pool,
	}
	for {
		item, ok := pool.next()
		if !ok {
			break
		}
		handle.item = item

		pool.task(&handle)
		handle.Sync()
	}
	atomic.AddInt32(&pool.numWorkers, -1)

	pool.wg.Done()
}
