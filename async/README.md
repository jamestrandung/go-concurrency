# Async

## Why you want to use this package
Package `async` simplifies the implementation of orchestration patterns for concurrent systems. It is similar to 
Java Future or JS Promise, which makes life much easier when dealing with asynchronous operation and concurrent 
processing. Golang is excellent in terms of parallel programming. However, dealing with goroutines and channels 
could be a big headache when business logic gets complicated. Wrapping them into higher-level functions improves
code readability significantly and makes it easier for engineers to reason about the system's behaviours.
                                                                                              
Currently, this package includes:

* Asynchronous tasks with cancellations, context propagation and state.
* Task chaining by using continuations.
* Fork/join pattern - running a batch of tasks in parallel and blocking until all finish.

## Concept
**Task** is a basic concept like `Future` in Java. You can create a `Task` using an executable function which takes 
in `context.Context`, then returns error and an optional result.

```go
task := NewTask(func(context.Context) (animal, error) {
    // run the job
    return res, err
})

silentTask := NewSilentTask(func(context.Context) error {
    // run the job
    return err
})
```

### Get the result
The function will be executed asynchronously. You can query whether it's completed by calling `task.State()`, which 
is a non-blocking function. Alternative, you can wait for the response using `task.Outcome()` or `silentTask.Wait()`, 
which will block the execution until the task is done. These functions are quite similar to the equivalents in Java
`Future.isDone()` or `Future.get()`.

### Cancelling
There could be case that we don't care about the result anymore some time after execution. In this case, a task can 
be aborted by invoking `task.Cancel()`.

### Chaining
To have a follow-up action after a task is done, you can use the provided family of `Continue` functions. This could 
be very useful to create a chain of processing, or to have a teardown process at the end of a task.

### Fork join
`ForkJoin` is meant for running multiple subtasks concurrently. They could be different parts of the main task which 
can be executed independently. The following code example illustrates how you can send files to S3 concurrently with 
a few lines of code.

```go
func uploadFilesConcurrently(files []string) {
    var tasks []Task[string]
    for _, file := range files {
        f := file
        
        tasks = append(tasks, NewTask(func(ctx context.Context) (string, error) {
            return upload(ctx, f)
        }))
    }

    ForkJoin(context.Background(), tasks)
}

func upload(ctx context.Context, file string) (string, error){
    // do file uploading
    return "", nil
}
```



