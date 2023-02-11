# Worker pool

Similar to a thread pool in Java, a worker pool is a fixed-size pool of goroutines that can execute any kind of tasks
as they come. When the number of tasks grows larger than the number of available workers, new tasks will be pushed into
a waiting queue to be executed later.

Our implementation is an upgrade/retrofit of the popular https://github.com/gammazero/workerpool repo to match the
behaviors of other features in this library.

```go
wp := NewWorkerPool(
    WithMaxSize(5), 
	WithBurst(10, 5), 
)
defer wp.Stop()

task := NewTask(func(context.Context) (animal, error) {
    // run the job
    return res, err
})

ctxWithTimeout, cancel := context.WithTimeout(context.Background(), 2*time.Second)
defer cancel()

// The given context will be used to execute the task
wp.Submit(ctxWithTimeout, task)

// Block and wait for the outcome
result, err := task.Outcome()
```