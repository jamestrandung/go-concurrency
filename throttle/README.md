# Throttle
Sometimes you don't really care about the concurrency level but just want to execute the tasks at a particular rate.
The `Throttle` function would come in handy in this case.

```go
// Throttle starts the given tasks at the specified rate. Clients can use
// the output task to block and wait for the tasks to complete if they want.
func Throttle[T async.SilentTask](tasks []T, rateLimit int, every time.Duration) async.SilentTask
```

For example, if you want to send 4 files every 2 seconds, the `Throttle` function will start a task every 0.5 second.