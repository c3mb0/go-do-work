package gdw

import (
	"sync/atomic"

	"github.com/eapache/channels"
)

// Rebel is used to define a goroutine pool whose purpose is to execute fire-and-forget jobs.
type Rebel struct {
	jobQueue   *channels.InfiniteChannel
	limiter    *channels.ResizableChannel
	queueDepth int64
}

func RebelPool(size int) *Rebel {
	jobQueue := channels.NewInfiniteChannel()
	limiter := channels.NewResizableChannel()
	limiter.Resize(channels.BufferCap(size))
	rebel := &Rebel{
		jobQueue:   jobQueue,
		limiter:    limiter,
		queueDepth: 0,
	}

	go func() {
		jobQueueOut := jobQueue.Out()
		limiterIn := limiter.In()
		limiterOut := limiter.Out()
		for jobs := range jobQueueOut {
			for _, job := range jobs.([]Job) {
				limiterIn <- true
				atomic.AddInt64(&rebel.queueDepth, -1)
				go func(j Job) {
					j.DoWork()
					<-limiterOut
				}(job)
			}
		}
	}()

	return rebel
}

func (r *Rebel) SetPoolSize(size int) {
	r.limiter.Resize(channels.BufferCap(size))
}

func (r *Rebel) GetPoolSize() int {
	return int(r.limiter.Cap())
}

func (r *Rebel) GetQueueDepth() int {
	return int(atomic.LoadInt64(&r.queueDepth))
}

func (r *Rebel) Add(job Job, amount int) {
	atomic.AddInt64(&r.queueDepth, int64(amount))
	jobs := make([]Job, amount)
	for i := 0; i < amount; i++ {
		jobs[i] = job
	}
	r.jobQueue.In() <- jobs
}

func (r *Rebel) Close() {
	r.jobQueue.Close()
	r.limiter.Close()
}
