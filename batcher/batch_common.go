// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package batcher

import "context"

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
}
