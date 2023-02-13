// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package batcher

import (
	"context"
	"sync"
	"time"
)

// iBatcher is a batch processor which is suitable for sitting in the back to accumulate
// tasks and then execute all in one go.
type iBatcher interface {
	// Size returns the length of the pending queue.
	Size() int
	// BuyTicket returns a context containing a ticket that can be used to auto process
	// when all ticket owners have arrived.
	BuyTicket(ctx context.Context) context.Context
	// DiscardTicket discards the issued ticket in the given context to allow other ticket
	// owners to get auto processed earlier.
	DiscardTicket(ctx context.Context)
	// Process executes all pending tasks in one go.
	Process(ctx context.Context)
	// Shutdown notifies this batch processor to complete its work gracefully. Future calls
	// to Append will return an error immediately and Process will be a no-op. This is a
	// blocking call which will wait up to the configured amount of time for the last batch
	// to complete.
	Shutdown()

	doProcess(ctx context.Context, isShuttingDown bool, toProcessBatchID uint64)
}

type baseBatcher struct {
	itself iBatcher
	sync.RWMutex
	*batcherConfigs
	isActive bool
	batchID  uint64
}

func (b *baseBatcher) isPeriodicAutoProcessingConfigured() bool {
	return b.autoProcessInterval > 0
}

func (b *baseBatcher) BuyTicket(ctx context.Context) context.Context {
	b.Lock()
	defer b.Unlock()

	return b.ticketBooth.sellTicket(ctx)
}

func (b *baseBatcher) DiscardTicket(ctx context.Context) {
	b.Lock()
	defer b.Unlock()

	b.ticketBooth.discardTicket(ctx)
}

func (b *baseBatcher) Process(ctx context.Context) {
	b.Lock()
	defer b.Unlock()

	b.itself.doProcess(ctx, false, b.batchID)
}

func (b *baseBatcher) Shutdown() {
	b.Lock()
	defer b.Unlock()

	ctx := context.Background()
	if b.shutdownGraceDuration > 0 {
		ctxWithTimeout, cancel := context.WithTimeout(ctx, b.shutdownGraceDuration)
		defer cancel()

		ctx = ctxWithTimeout
	}

	b.itself.doProcess(ctx, true, b.batchID)

	b.isActive = false
}

func (b *baseBatcher) spawnGoroutineToAutoProcessPeriodically() {
	if !b.isPeriodicAutoProcessingConfigured() {
		return
	}

	go func() {
		for {
			curBatchId := func() uint64 {
				b.RLock()
				defer b.RUnlock()

				return b.batchID
			}()

			<-time.After(b.autoProcessInterval)

			func() {
				b.Lock()
				defer b.Unlock()

				b.itself.doProcess(context.Background(), false, curBatchId)
			}()

			shouldBreak := func() bool {
				b.RLock()
				defer b.RUnlock()

				return !b.isActive
			}()

			if shouldBreak {
				return
			}
		}
	}()
}
