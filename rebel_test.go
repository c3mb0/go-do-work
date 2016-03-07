package gdw

import (
	"testing"
	"time"
)

func TestRebelPool(t *testing.T) {
	test := &adder{count: 0}
	pool := RebelPool(3)
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
	time.Sleep(10 * time.Second)
}
