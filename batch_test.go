package gdw

import (
	"fmt"
	"testing"
)

func TestBatch(t *testing.T) {
	test := adder{count: 0}
	pool := NewWorkerPool(3)
	defer pool.Close()
	batch1 := pool.NewTempBatch()
	batch2 := pool.NewTempBatch()
	pool.NewBatch("my batch")
	defer batch1.Clean()
	defer batch2.Clean()
	defer pool.CleanBatch("my batch")
	batch1.Add(test, 5)
	batch2.Add(test, 10)
	batch3, _ := pool.LoadBatch("my batch")
	batch3.Add(test, 4)
	batch1.Wait()
	fmt.Println("batch 1 done")
	batch2.Wait()
	fmt.Println("batch 2 done")
	fmt.Println(pool.GetQueueDepth())
	pool.Wait()
}
