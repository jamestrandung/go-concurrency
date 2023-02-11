// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package promising

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

// IOutcome is a promise for a future outcome.
//go:generate mockery --name IOutcome --case underscore --inpackage
type IOutcome[V any] interface {
	// cancel this outcome if it has not completed yet.
	cancel()
	// complete this outcome with the given result if it has not completed yet.
	complete(val V, err error)
	// Get blocks and waits for the outcome to complete and then returns the result.
	Get(ctx context.Context) (V, error)
}

// outcome is a promise for a future outcome.
type outcome[K comparable, V any] struct {
	key        K
	done       chan struct{}
	isComplete bool
	value      V
	err        error
}

// newPendingOutcome returns a pending outcome.
func newPendingOutcome[K comparable, V any](key K) *outcome[K, V] {
	return &outcome[K, V]{
		key:  key,
		done: make(chan struct{}),
	}
}

// newErrorOutcome returns an outcome completed with the given error.
func newErrorOutcome[K comparable, V any](key K, err error) *outcome[K, V] {
	done := make(chan struct{})
	close(done)

	return &outcome[K, V]{
		key:        key,
		done:       done,
		isComplete: true,
		err:        err,
	}
}

// complete this outcome with the given result if it has not completed yet.
func (o *outcome[K, V]) complete(val V, err error) {
	if o.isComplete {
		return
	}

	o.isComplete = true
	o.value = val
	o.err = err
	close(o.done)
}

// cancel this outcome if it has not completed yet.
func (o *outcome[K, V]) cancel() {
	if o.isComplete {
		return
	}

	o.isComplete = true
	o.err = errors.Wrap(ErrNoResultForGivenKey, fmt.Sprintf("key(%v)", o.key))
	close(o.done)
}

// Get blocks and waits for the outcome to complete and then returns the result.
func (o *outcome[K, V]) Get(ctx context.Context) (V, error) {
	select {
	case <-ctx.Done():
		var temp V
		return temp, ctx.Err()
	case <-o.done:
		return o.value, o.err
	}
}
