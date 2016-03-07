package gdw

import (
	"sync"

	"github.com/eapache/channels"
)

// Worker is used to define a goroutine pool whose results and execution are of interest, thus awaitable through WaitGroup.
type Worker struct {
	jobQueue *channels.InfiniteChannel
	limiter  *channels.ResizableChannel
	wg       sync.WaitGroup
}

func WorkerPool(size int) *Worker {
	jobQueue := channels.NewInfiniteChannel()
	limiter := channels.NewResizableChannel()
	limiter.Resize(channels.BufferCap(size))
	worker := &Worker{
		jobQueue: jobQueue,
		limiter:  limiter,
	}

	go func() {
		jobQueueOut := jobQueue.Out()
		limiterIn := limiter.In()
		limiterOut := limiter.Out()
		for jobs := range jobQueueOut {
			for _, job := range jobs.([]Job) {
				limiterIn <- true
				go func(j Job) {
					defer worker.wg.Done()
					j.DoWork()
					<-limiterOut
				}(job)
			}
		}
	}()

	return worker
}

func (w *Worker) SetPoolSize(amount int) {
	w.limiter.Resize(channels.BufferCap(amount))
}

func (w *Worker) GetPoolSize() int {
	return int(w.limiter.Cap())
}

func (w *Worker) Add(job Job, amount int) {
	w.wg.Add(amount)
	jobs := make([]Job, amount)
	for i := 0; i < amount; i++ {
		jobs[i] = job
	}
	w.jobQueue.In() <- jobs
}

func (w *Worker) Wait() {
	w.wg.Wait()
}

func (w *Worker) Close() {
	w.jobQueue.Close()
	w.limiter.Close()
}
