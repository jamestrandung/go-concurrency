// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package promising

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
)

// ICache is an in-memory cache of outcome associated with specific keys.
//go:generate mockery --name ICache --case underscore --inpackage
type ICache[K comparable, V any] interface {
	// Destroy marks this cache as destroyed and cancel all pending outcomes.
	Destroy()
	// GetOutcome returns the outcome associated with the given key.
	GetOutcome(key K) IOutcome[V]
	// Complete the outcome associated with the given key using the given result.
	Complete(key K, val V, err error)
}

// cache is an in-memory cache of outcome associated with specific keys.
type cache[K comparable, V any] struct {
	sync.Mutex
	isDestroyed bool
	outcomes    map[K]IOutcome[V]
}

// NewCache ...
func NewCache[K comparable, V any]() ICache[K, V] {
	return &cache[K, V]{
		outcomes: make(map[K]IOutcome[V]),
	}
}

// Destroy marks this cache as destroyed and cancel all pending outcomes.
func (c *cache[K, V]) Destroy() {
	c.Lock()
	defer c.Unlock()

	c.isDestroyed = true

	// Cancel all pending outcomes
	for _, o := range c.outcomes {
		o.cancel()
	}
}

// GetOutcome returns the outcome associated with the given key.
func (c *cache[K, V]) GetOutcome(key K) IOutcome[V] {
	c.Lock()
	defer c.Unlock()

	if o, ok := c.outcomes[key]; ok {
		return o
	}

	if c.isDestroyed {
		return newErrorOutcome[K, V](
			key,
			errors.Wrap(ErrNoMoreResultIncoming, fmt.Sprintf("key: %v", key)),
		)
	}

	o := newPendingOutcome[K, V](key)
	c.outcomes[key] = o

	return o
}

// Complete the outcome associated with the given key using the given result.
func (c *cache[K, V]) Complete(key K, val V, err error) {
	o := c.GetOutcome(key)
	o.complete(val, err)
}
