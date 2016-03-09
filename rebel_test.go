package gdw

import (
	"fmt"
	"testing"
	"time"
)

func TestRebelPool(t *testing.T) {
	test := adder{count: 0}
	pool := NewRebelPool(3)
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
	time.Sleep(10 * time.Second)
}
