# Spread

Instead of starting all tasks at once with `ForkJoin`, you can also spread the starting points of your tasks evenly
within a certain duration using the `Spread` function.

```go
// Spread evenly starts the given tasks within the specified duration. Clients can
// use the output task to block and wait for the tasks to complete if they want.
func Spread[T async.SilentTask](tasks []T, within time.Duration) async.SilentTask
```

For example, if you want to send 50 files within 10 seconds, the `Spread` function would start a task every 0.2s.