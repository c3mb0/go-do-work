package gdw

import (
	"sync/atomic"
	"testing"
	"time"
)

type adder struct {
	count uint32
}

func (a *adder) DoWork() {
	atomic.AddUint32(&a.count, 1)
	time.Sleep(1 * time.Second)
}

func TestWorkerPool(t *testing.T) {
	test := &adder{count: 0}
	pool := WorkerPool(3)
	defer pool.Close()
	pool.Add(test, 10)
	time.Sleep(1 * time.Second)
	pool.GetQueueDepth()
	time.Sleep(1 * time.Second)
	pool.SetPoolSize(5)
	time.Sleep(1 * time.Second)
	pool.GetQueueDepth()
	time.Sleep(1 * time.Second)
	pool.SetPoolSize(2)
	pool.Wait()
}
