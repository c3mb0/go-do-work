package gdw

type Batch struct {
	worker *Worker
	name   string
}

func (b *Batch) Add(job Job, amount int) {
	w := b.worker
	index := w.indexMap[b.name]
	w.wgSlice[index].Add(amount)
	w.add(job, amount, index)
}

func (b *Batch) AddOne(job Job) {
	w := b.worker
	index := w.indexMap[b.name]
	w.wgSlice[index].Add(1)
	w.addOne(job, index)
}

func (b *Batch) Wait() {
	b.worker.wait(b.worker.indexMap[b.name])
}

func (b *Batch) Remove() error {
	return b.worker.RemoveBatch(b.name)
}
