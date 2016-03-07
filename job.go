package gdw

// Job interface defines a method through which gdw can execute requested jobs.
type Job interface {
	DoWork()
}
