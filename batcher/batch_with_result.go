// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package batcher

import (
	"context"
	"github.com/jamestrandung/go-concurrency/v2/async"
)

type batchEntry[P any] struct {
	id      uint64
	payload P // Will be used as input when the batch is processed
}

type pendingBatch[P any] struct {
	// size of a batch = the number of requests arrived which
	// may be larger than the number of entries due to the
	// single-flight feature
	arrivedRequestSize  int
	entries             []batchEntry[P]
	singleFlightCache   map[any]uint64
	payloadKeyExtractor func(P) any
}

func newPendingBatch[P any](payloadKeyExtractor func(P) any) *pendingBatch[P] {
	return &pendingBatch[P]{
		singleFlightCache:   make(map[any]uint64),
		payloadKeyExtractor: payloadKeyExtractor,
	}
}

func (pb *pendingBatch[P]) size() int {
	return pb.arrivedRequestSize
}

func (pb *pendingBatch[P]) incrementBatchSize() {
	pb.arrivedRequestSize = pb.arrivedRequestSize + 1
}

func (pb *pendingBatch[P]) attachToAnExistingEntry(payload P) (uint64, bool) {
	if pb.payloadKeyExtractor == nil {
		return 0, false
	}

	payloadKey := pb.payloadKeyExtractor(payload)
	entryID, ok := pb.singleFlightCache[payloadKey]
	if ok {
		pb.incrementBatchSize()
	}

	return entryID, ok
}

func (pb *pendingBatch[P]) append(id uint64, payload P) {
	pb.entries = append(
		pb.entries, batchEntry[P]{
			id:      id,
			payload: payload,
		},
	)

	pb.incrementBatchSize()

	if pb.payloadKeyExtractor == nil {
		return
	}

	payloadKey := pb.payloadKeyExtractor(payload)
	pb.singleFlightCache[payloadKey] = id
}

// Batcher is a batch processor which is suitable for sitting in the back to receive tasks
// from callers to execute in one go and then return individual result to each caller.
//
//go:generate mockery --name Batcher --case underscore --inpackage
type Batcher[P any, T any] interface {
	iBatcher
	// Append adds a new payload to the batch and returns a task for that particular payload.
	// Clients MUST execute the returned task before blocking and waiting for it to complete
	// to extract result.
	Append(ctx context.Context, payload P) async.Task[T]
}

type batcher[P any, T any] struct {
	*baseBatcher
	lastID              uint64                   // The current entry ID
	pending             *pendingBatch[P]         // The task queue to be executed in one batch
	batchExecutor       async.Task[map[uint64]T] // The current batch executor
	payloadKeyExtractor func(P) any              // The key extract handling incoming payloads to support single-flight
	batch               chan *pendingBatch[P]    // The channel to submit a batch to be processed by the above executor
	processFn           func([]P) ([]T, error)   // The func which will be executed to process one batch of tasks
}

// NewBatcher returns a new Batcher
func NewBatcher[P any, T any](
	processFn func([]P) ([]T, error),
	payloadKeyExtractor func(P) any,
	options ...BatcherOption,
) Batcher[P, T] {
	configs := &batcherConfigs{
		ticketBooth: noOpTicketBooth{},
	}

	for _, o := range options {
		o(configs)
	}

	b := &batcher[P, T]{
		baseBatcher: &baseBatcher{
			batcherConfigs: configs,
			isActive:       true,
		},
		pending:   newPendingBatch(payloadKeyExtractor),
		batch:     make(chan *pendingBatch[P], 1),
		processFn: processFn,
	}

	b.itself = b

	b.spawnGoroutineToAutoProcessPeriodically()

	return b
}

func (b *batcher[P, T]) Append(ctx context.Context, payload P) async.Task[T] {
	b.Lock()
	defer b.Unlock()

	if !b.isActive {
		var temp T
		return async.Completed[T](temp, ErrBatchProcessorNotActive)
	}

	// Make sure we have a batch executor
	curBatchExecutor := b.batchExecutor
	if curBatchExecutor == nil {
		b.batchExecutor = b.createBatchExecutor()
		curBatchExecutor = b.batchExecutor
	}

	// Reuse existing entry if possible
	if existingEntryID, ok := b.pending.attachToAnExistingEntry(payload); ok {
		return async.NewTask[T](
			func(ctx context.Context) (T, error) {
				batchResult, err := curBatchExecutor.Outcome()
				if err != nil {
					var temp T
					return temp, err
				}

				return batchResult[existingEntryID], nil
			},
		)
	}

	// Generate a new ID for the current batch entry
	b.lastID = b.lastID + 1
	id := b.lastID

	// Add to the task queue
	b.pending.append(id, payload)

	// Auto process if configured and reached the threshold
	if b.shouldAutoProcess(ctx) {
		curBatchId := b.batchID

		go func() {
			b.Lock()
			defer b.Unlock()

			b.doProcess(context.Background(), false, curBatchId)
		}()
	}

	// Return a task for caller to execute themselves to block
	// & wait using their own context.
	return async.NewTask[T](
		func(ctx context.Context) (T, error) {
			batchResult, err := curBatchExecutor.Outcome()
			if err != nil {
				var temp T
				return temp, err
			}

			return batchResult[id], nil
		},
	)
}

func (b *batcher[P, T]) Size() int {
	b.RLock()
	defer b.RUnlock()

	return b.pending.size()
}

func (b *batcher[P, T]) shouldAutoProcess(ctx context.Context) bool {
	haveAllClientsArrived := b.ticketBooth.submitTicket(ctx)
	shouldAutoProcessBySize := b.autoProcessSize > 0 && b.pending.size() == b.autoProcessSize

	return haveAllClientsArrived || shouldAutoProcessBySize
}

func (b *batcher[P, T]) doProcess(ctx context.Context, isShuttingDown bool, toProcessBatchID uint64) {
	if b.batchID != toProcessBatchID {
		return
	}

	if b.pending.size() == 0 {
		if isShuttingDown {
			close(b.batch)
		}

		return
	}

	// Capture pending tasks and reset the queue
	pending := b.pending
	b.pending = newPendingBatch[P](b.payloadKeyExtractor)

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
	b.batchID = b.batchID + 1
}

// createBatchExecutor creates an executor for one batch of tasks.
func (b *batcher[P, T]) createBatchExecutor() async.Task[map[uint64]T] {
	return async.NewTask[map[uint64]T](
		func(context.Context) (map[uint64]T, error) {
			// Block here until a batch is submitted to be processed
			pending := <-b.batch

			// Prepare the input for the batch process call
			input := make([]P, len(pending.entries))
			for idx, entry := range pending.entries {
				input[idx] = entry.payload
			}

			// Process the batch
			result, err := b.processFn(input)
			if err != nil {
				return nil, err
			}

			// Map the result back to individual entry
			m := make(map[uint64]T)
			for i, res := range result {
				id := pending.entries[i].id
				m[id] = res
			}

			return m, nil
		},
	)
}
