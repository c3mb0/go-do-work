# go-do-work

`gdw` makes use of [eapache's delightfully clever channels package](https://github.com/eapache/channels) in order to provide dynamically resizable pools of goroutines which can queue an infinite number of jobs.

## Installation

`go get github.com/c3mb0/go-do-work`

## Types of Pools

There are currently 2 types of pools in `gdw`: `Worker` and `Rebel`. Their internal mechanics of operation are the same except for jobs queued in a `WorkerPool` being waitable. This allows for separation of concerns, namely for jobs whose results and/or execution are of interest and jobs which are of fire-and-forget nature. You can safely mix them without affecting one and other.

## Usage

Any object that implements the `Job` interface is eligible to be queued and executed by `gdw`. There is, however, a __big__ difference between queuing an object and an object *pointer*.

Consider the following example:
```
type adder struct {
	count int
}

func (a adder) DoWork() {
	a.count++
	fmt.Print(a.count, " ")
}

func main() {
	test := adder{count: 0}
	pool := gdw.WorkerPool(2)
	defer pool.Close()
	pool.Add(test, 5)
	pool.Wait()
	fmt.Println()
}
```
Here we create a new `WorkerPool` with a pool size of 2. We then queue 5 `test` jobs. The resulting output is:
```
1 1 1 1 1
```
Let's queue an object pointer instead of an object:
```
type adder struct {
	count int
}

func (a *adder) DoWork() {
	a.count++
	fmt.Print(a.count, " ")
}

func main() {
	test := &adder{count: 0}
	pool := gdw.WorkerPool(2)
	defer pool.Close()
	pool.Add(test, 5)
	pool.Wait()
	fmt.Println()
}
```
The resulting output is:
```
1 2 3 4 5
```
When you queue an object, each goroutine in the pool works on a copy of the object provided. On the other hand, when you queue an object pointer, all goroutines work on the same object. These approaches both have their use cases, but keep in mind that the latter approach needs to be thread-safe. Thus, the correct implementation would be:
```
type adder struct {
	count uint32
}

func (a *adder) DoWork() {
	atomic.AddUint32(&a.count, 1)
	fmt.Print(atomic.LoadUint32(&a.count), " ")
}

func main() {
	test := &adder{count: 0}
	pool := gdw.WorkerPool(2)
	defer pool.Close()
	pool.Add(test, 5)
	pool.Wait()
	fmt.Println()
}
```

## Pool Size and Safety

You can safely increase or decrease a pool's size at runtime without losing already queued data or shutting down already running goroutines. The only caveat is that you cannot set the pool size to 0. Details can be found [here](https://github.com/eapache/channels/issues/1).

The following example demonstrates pool resizing in action:
```
type adder struct {
	count int
}

func (a adder) DoWork() {
	a.count++
	fmt.Print(a.count, " ")
	time.Sleep(2 * time.Second)
}

func main() {
	test := adder{count: 0}
	pool := gdw.WorkerPool(3)
	defer pool.Close()
	pool.Add(test, 5)
	time.Sleep(1 * time.Second)
	pool.SetPoolSize(1)
	fmt.Print("\n", pool.GetQueueDepth(), "\n")
	pool.Wait()
	fmt.Println()
}
```
Check the output for some magic!

## Collecting Results

If you would like to get some results back from your jobs, the most practical approach is to slip in a channel to the object of interest:
```
type adder struct {
	count  int
	result chan int
}

func (a adder) DoWork() {
	a.count++
	a.result <- a.count
}

func main() {
	result := make(chan int)
	test := adder{
		count:  0,
		result: result,
	}
	pool := gdw.WorkerPool(3)
	defer pool.Close()
	pool.Add(test, 5)
	go func() {
		for res := range result {
			fmt.Print(res, " ")
		}
	}()
	pool.Wait()
	close(result) // close the result channel after the pool has completed
	fmt.Println()
}
```
This works for both object and object pointer jobs; each batch will share the same channel.
