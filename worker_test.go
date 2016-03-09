package gdw

import (
	"fmt"
	"testing"
	"time"
)

type adder struct {
	count uint32
}

func (a adder) DoWork() {
	a.count++
	time.Sleep(1 * time.Second)
}

func TestWorkerPool(t *testing.T) {
	test := adder{count: 0}
	pool := NewWorkerPool(3)
	defer pool.Close()
	pool.Add(test, 10)
	time.Sleep(1 * time.Second)
	fmt.Println(pool.GetQueueDepth())
	time.Sleep(1 * time.Second)
	pool.SetPoolSize(5)
	time.Sleep(1 * time.Second)
	fmt.Println(pool.GetQueueDepth())
	time.Sleep(1 * time.Second)
	pool.SetPoolSize(2)
	fmt.Println(pool.GetPoolSize())
	pool.Wait()
}
