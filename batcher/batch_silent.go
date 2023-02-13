// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package batcher

import (
	"context"
	"errors"
	"github.com/jamestrandung/go-concurrency/async"
)

var (
	ErrBatchProcessorNotActive = errors.New("batch processor has already shut down")
)

type silentBatch[P any] []silentBatchEntry[P]

type silentBatchEntry[P any] struct {
	payload P // Will be used as input when the batch is processed
}

// SilentBatcher is a batch processor which is suitable for sitting in the back to accumulate
// tasks and then execute all in one go silently.
//go:generate mockery --name SilentBatcher --case underscore --inpackage
type SilentBatcher[P any] interface {
	iBatcher
	// Append adds a new payload to the batch and returns a task for that particular payload.
	// Clients MUST execute the returned task before blocking and waiting for it to complete
	// to extract result.
	Append(ctx context.Context, payload P) async.SilentTask
}

type silentBatcher[P any] struct {
	*baseBatcher
	pending       silentBatch[P]      // The task queue to be executed in one batch
	batchExecutor async.SilentTask    // The current batch executor
	batch         chan silentBatch[P] // The channel to submit a batch to be processed by the above executor
	processFn     func([]P) error     // The func which will be executed to process one batch of tasks
}

// NewSilentBatcher returns a new SilentBatcher
func NewSilentBatcher[P any](processFn func([]P) error, options ...BatcherOption) SilentBatcher[P] {
	b := &silentBatcher[P]{
		baseBatcher: &baseBatcher{
			batcherConfigs: &batcherConfigs{
				ticketBooth: noOpTicketBooth{},
			},
			isActive: true,
		},
		pending:   silentBatch[P]{},
		batch:     make(chan silentBatch[P], 1),
		processFn: processFn,
	}

	for _, o := range options {
		o(b.batcherConfigs)
	}

	b.itself = b

	b.spawnGoroutineToAutoProcessPeriodically()

	return b
}

func (b *silentBatcher[P]) Append(ctx context.Context, payload P) async.SilentTask {
	b.Lock()
	defer b.Unlock()

	if !b.isActive {
		return async.Completed(struct{}{}, ErrBatchProcessorNotActive)
	}

	// Make sure we have a batch executor
	curBatchExecutor := b.batchExecutor
	if curBatchExecutor == nil {
		b.batchExecutor = b.createBatchExecutor()
		curBatchExecutor = b.batchExecutor
	}

	// Add to the task queue
	b.pending = append(
		b.pending, silentBatchEntry[P]{
			payload: payload,
		},
	)

	// Auto process if configured and reached the threshold
	if b.shouldAutoProcess(ctx) {
		curBatchId := b.batchID

		go func() {
			b.Lock()
			defer b.Unlock()

			b.doProcess(context.Background(), false, curBatchId)
		}()
	}

	return async.NewSilentTask(
		func(ctx context.Context) error {
			return curBatchExecutor.Error()
		},
	)
}

func (b *silentBatcher[P]) Size() int {
	b.RLock()
	defer b.RUnlock()

	return len(b.pending)
}

func (b *silentBatcher[P]) shouldAutoProcess(ctx context.Context) bool {
	shouldAutoProcess := b.autoProcessSize > 0 && len(b.pending) == b.autoProcessSize
	if shouldAutoProcess {
		return true
	}

	return b.ticketBooth.submitTicket(ctx)
}

func (b *silentBatcher[P]) doProcess(ctx context.Context, isShuttingDown bool, toProcessBatchID uint64) {
	if b.batchID != toProcessBatchID {
		return
	}

	if len(b.pending) == 0 {
		if isShuttingDown {
			close(b.batch)
		}

		return
	}

	// Capture pending tasks and reset the queue
	pending := b.pending
	b.pending = silentBatch[P]{}

	// Run the current batch using the existing executor
	b.batch <- pending
	b.batchExecutor.Execute(ctx)

	// Block and wait for the last batch to complete on shutting down
	if isShuttingDown {
		b.batchExecutor.Wait()
		return
	}

	// Prepare a new executor
	b.batchExecutor = b.createBatchExecutor()

	// Increment batch ID to stop the timer from processing old batch
	b.batchID += 1
}

// createBatchExecutor creates an executor for one batch of tasks.
func (b *silentBatcher[P]) createBatchExecutor() async.SilentTask {
	return async.NewSilentTask(
		func(context.Context) error {
			// Block here until a batch is submitted to be processed
			pending := <-b.batch

			// Prepare the input for the batch process call
			input := make([]P, len(pending))
			for idx, entry := range pending {
				input[idx] = entry.payload
			}

			return b.processFn(input)
		},
	)
}
