# REPEAT

In cases where you need to repeat a background task on a pre-determined interval, `Repeat` is your friend. The 
returned `SilentTask` can then be used to cancel the repeating task at any time.

```go
// Repeat executes the given action asynchronously on a pre-determined interval. The repeating
// process can be stopped by cancelling the output task. All errors and panics must be handled
// inside the action if callers want the process to continue. Otherwise, the repeat will stop.
func Repeat(interval time.Duration, action async.SilentWork) async.SilentTask
```