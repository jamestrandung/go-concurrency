# Batch

Instead of executing a task immediately whenever you receive an input, sometimes, it might be more efficient to
create a batch of inputs and process all in one go.

```go
out := make(chan int, taskCount)

processBatch := func(nums []int) error {
    for _, number := range nums {
        out <- number * 10
    }
    
    return nil
}

// Auto process batch every 100ms
periodicBatcher := NewBatcher(
    processBatch,
    WithAutoProcessInterval(100*time.Millisecond),
)

// Auto process batch when pending queue reaches 10
sizeBatcher := NewBatcher(
    processBatch,
    WithAutoProcessSize(10),
)

// Auto process batch every 100ms or when pending queue reaches 10
periodicSizeBatcher := NewBatcher(
    processBatch,
    WithAutoProcessInterval(100*time.Millisecond),
    WithAutoProcessSize(10),
)

// Auto process batch when pending queue reaches 10
manualBatcher := NewBatcher(
    processBatch,
)

manualBatcher.Process()
```

See `batch_test.go` for more detailed examples on how to use this feature.