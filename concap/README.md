# Concurrency cap

`ForkJoin` is not suitable when the number of tasks is huge. In this scenario, the number of concurrent goroutines
would likely overwhelm a node and consume too much CPU resources. One solution is to put a cap on the max concurrency
level. `RunWithConcurrencyLevelC` and `RunWithConcurrencyLevelS` were created for this purpose. Internally, it's like
maintaining a fixed-size worker pool which aims to execute the given tasks as quickly as possible without violating
the given constraint.

```go
// RunWithConcurrencyLevelC runs the given tasks up to the max concurrency level. Clients can use the output
// task to block and wait for the tasks to complete if they want.
//
// Note: When `ctx` is cancelled, we spawn a new goroutine to cancel all remaining tasks in the given channel.
// To avoid memory leak, clients MUST make sure new tasks will eventually stop arriving once `ctx` is cancelled
// so that the new goroutine can return.
func RunWithConcurrencyLevelC[T async.SilentTask](concurrencyLevel int, tasks <-chan T) async.SilentTask

// RunWithConcurrencyLevelS runs the given tasks up to the max concurrency level. Clients
// can use the output task to block and wait for the tasks to complete if they want.
func RunWithConcurrencyLevelS[T async.SilentTask](concurrencyLevel int, tasks []T) async.SilentTask
```