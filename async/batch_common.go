package async

import "context"

// iBatcher is a batch processor which is suitable for sitting in the back to accumulate
// tasks and then execute all in one go.
type iBatcher interface {
	// Size returns the length of the pending queue.
	Size() int
	// Process executes all pending tasks in one go.
	Process(ctx context.Context)
	// Shutdown notifies this batch processor to complete its work gracefully. Future calls
	// to Append will return an error immediately and Process will be a no-op. This is a
	// blocking call which will wait up to the configured amount of time for the last batch
	// to complete.
	Shutdown()
}
