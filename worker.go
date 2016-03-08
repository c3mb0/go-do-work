package gdw

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/eapache/channels"
)

// Worker is used to define a goroutine pool whose results and/or execution are of interest, thus awaitable through WaitGroup.
type Worker struct {
	jobQueue   *channels.InfiniteChannel
	limiter    *channels.ResizableChannel
	queueDepth int64
	wgSlice    []sync.WaitGroup
	indexMap   map[string]int
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
	worker.wgSlice = append(worker.wgSlice, sync.WaitGroup{})
	// worker.indexMap["gdw_root_wg"] = 0

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
					defer worker.wgSlice[0].Done()
					j.DoWork()
					<-limiterOut
				}(jt)

			case []Job:
				for _, job := range jt {
					limiterIn <- true
					atomic.AddInt64(&worker.queueDepth, -1)
					go func(j Job) {
						defer worker.wgSlice[0].Done()
						j.DoWork()
						<-limiterOut
					}(job)
				}

			case *batchedJob:
				limiterIn <- true
				atomic.AddInt64(&worker.queueDepth, -1)
				go func(bj *batchedJob) {
					defer worker.wgSlice[bj.index].Done()
					bj.batched.DoWork()
					<-limiterOut
				}(jt)

			case []*batchedJob:
				for _, job := range jt {
					limiterIn <- true
					atomic.AddInt64(&worker.queueDepth, -1)
					go func(bj *batchedJob) {
						defer worker.wgSlice[bj.index].Done()
						bj.batched.DoWork()
						<-limiterOut
					}(job)
				}
			}

		}
	}()

	return worker
}

func (w *Worker) NewBatch(name string) *Batch {
	w.indexMap[name] = len(w.wgSlice)
	w.wgSlice = append(w.wgSlice, sync.WaitGroup{})
	return &Batch{
		worker: w,
		name:   name,
	}
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
	w.add(job, amount, 0)
}

func (w *Worker) add(job Job, amount int, index int) {
	w.wgSlice[0].Add(amount)
	atomic.AddInt64(&w.queueDepth, int64(amount))
	switch index {
	case 0:
		jobs := make([]Job, amount)
		for i := 0; i < amount; i++ {
			jobs[i] = job
		}
		w.jobQueue.In() <- jobs
	default:
		jobs := make([]*batchedJob, amount)
		for i := 0; i < amount; i++ {
			bj := &batchedJob{
				batched: job,
				index:   index,
			}
			jobs[i] = bj
		}
		w.jobQueue.In() <- jobs
	}
}

func (w *Worker) AddOne(job Job) {
	w.addOne(job, 0)
}

func (w *Worker) addOne(job Job, index int) {
	w.wgSlice[0].Add(1)
	atomic.AddInt64(&w.queueDepth, 1)
	switch index {
	case 0:
		w.jobQueue.In() <- job
	default:
		bj := &batchedJob{
			batched: job,
			index:   index,
		}
		w.jobQueue.In() <- bj
	}
}

func (w *Worker) Wait() {
	w.wait(0)
}

func (w *Worker) WaitBatch(batch string) error {
	i, ok := w.indexMap[batch]
	if !ok {
		return fmt.Errorf("No batch named %s exists.", batch)
	}
	w.wait(i)
	return nil
}

func (w *Worker) wait(index int) {
	w.wgSlice[index].Wait()
}

func (w *Worker) RemoveBatch(batch string) error {
	i, ok := w.indexMap[batch]
	if !ok {
		return fmt.Errorf("No batch named %s exists.", batch)
	}
	w.wgSlice = append(w.wgSlice[:i], w.wgSlice[i+1:]...)
	delete(w.indexMap, batch)
	return nil
}

func (w *Worker) Close() {
	w.jobQueue.Close()
	w.limiter.Close()
}
