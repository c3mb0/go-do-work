package gdw

import (
	"sync"
	"sync/atomic"

	"github.com/eapache/channels"
)

// Worker is used to define a goroutine pool whose results and/or execution are of interest, thus awaitable through WaitGroup.
type Worker struct {
	jobQueue   *channels.InfiniteChannel
	limiter    *channels.ResizableChannel
	queueDepth int64
	wg         sync.WaitGroup
}

func WorkerPool(size int) *Worker {
	jobQueue := channels.NewInfiniteChannel()
	limiter := channels.NewResizableChannel()
	limiter.Resize(channels.BufferCap(size))
	worker := &Worker{
		jobQueue:   jobQueue,
		limiter:    limiter,
		queueDepth: 0,
	}

	go func() {
		jobQueueOut := jobQueue.Out()
		limiterIn := limiter.In()
		limiterOut := limiter.Out()
		for jobs := range jobQueueOut {
			switch jt := jobs.(type) {
			case Job:
				limiterIn <- true
				atomic.AddInt64(&worker.queueDepth, -1)
				go func(j Job) {
					defer worker.wg.Done()
					j.DoWork()
					<-limiterOut
				}(jt)
			case []Job:
				for _, job := range jt {
					limiterIn <- true
					atomic.AddInt64(&worker.queueDepth, -1)
					go func(j Job) {
						defer worker.wg.Done()
						j.DoWork()
						<-limiterOut
					}(job)
				}
			}
		}
	}()

	return worker
}

func (w *Worker) SetPoolSize(size int) {
	w.limiter.Resize(channels.BufferCap(size))
}

func (w *Worker) GetPoolSize() int {
	return int(w.limiter.Cap())
}

func (w *Worker) GetQueueDepth() int {
	return int(atomic.LoadInt64(&w.queueDepth))
}

func (w *Worker) Add(job Job, amount int) {
	w.wg.Add(amount)
	atomic.AddInt64(&w.queueDepth, int64(amount))
	jobs := make([]Job, amount)
	for i := 0; i < amount; i++ {
		jobs[i] = job
	}
	w.jobQueue.In() <- jobs
}

func (w *Worker) AddOne(job Job) {
	w.wg.Add(1)
	atomic.AddInt64(&w.queueDepth, 1)
	w.jobQueue.In() <- job
}

func (w *Worker) Wait() {
	w.wg.Wait()
}

func (w *Worker) Close() {
	w.jobQueue.Close()
	w.limiter.Close()
}
