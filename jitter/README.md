# Jitter

Using jitters to avoid thundering herds is a popular technique. If you're not familiar with it, do
take a look at the article below.

http://highscalability.com/blog/2012/4/17/youtube-strategy-adding-jitter-isnt-a-bug.html

```go
// DoJitter adds a random jitter before executing doFn, then returns the jitter duration.
func DoJitter(doFn func(), maxJitterDurationInMilliseconds int) int

// AddJitterT adds a random jitter before executing the given Task.
func AddJitterT[T any](t async.Task[T], maxJitterDurationInMilliseconds int) async.Task[T]

// AddJitterST adds a random jitter before executing the given SilentTask.
func AddJitterST(t async.SilentTask, maxJitterDurationInMilliseconds int) async.SilentTask
```

See `jitter_test.go` for a detailed example on how to use this feature.