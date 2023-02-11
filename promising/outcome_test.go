// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package promising

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewPendingOutcome(t *testing.T) {
	o := newPendingOutcome[string, int]("abc")

	if !assert.NotNil(t, o) {
		return
	}

	assert.Equal(t, "abc", o.key)
	assert.NotNil(t, o.done)
	assert.False(t, o.isComplete)
	assert.Equal(t, 0, o.value)
	assert.Nil(t, o.err)

	assert.NotPanics(
		t, func() {
			close(o.done)
		},
		"channel should be closable",
	)
}

func TestNewErrorOutcome(t *testing.T) {
	o := newErrorOutcome[string, int]("abc", assert.AnError)

	if !assert.NotNil(t, o) {
		return
	}

	assert.Equal(t, "abc", o.key)
	assert.NotNil(t, o.done)
	assert.True(t, o.isComplete)
	assert.Equal(t, 0, o.value)
	assert.Equal(t, assert.AnError, o.err)

	assert.Panics(
		t, func() {
			close(o.done)
		},
		"channel should NOT be closable",
	)
}

func TestOutcome_Complete(t *testing.T) {
	o := newPendingOutcome[string, string]("abc")

	assert.Equal(t, "abc", o.key)
	assert.False(t, o.isComplete)
	assert.Equal(t, "", o.value)
	assert.Nil(t, o.err)

	o.complete("test", assert.AnError)

	assert.Equal(t, "abc", o.key)
	assert.True(t, o.isComplete)
	assert.Equal(t, "test", o.value)
	assert.Equal(t, assert.AnError, o.err)

	assert.Panics(
		t, func() {
			close(o.done)
		},
		"channel should NOT be closable",
	)
}

func TestOutcome_Cancel(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(t *testing.T)
	}{
		{
			desc: "pending outcome",
			test: func(t *testing.T) {
				o := newPendingOutcome[string, int]("abc")

				assert.Equal(t, "abc", o.key)
				assert.False(t, o.isComplete)
				assert.Equal(t, 0, o.value)
				assert.Nil(t, o.err)

				o.cancel()

				assert.Equal(t, "abc", o.key)
				assert.True(t, o.isComplete)
				assert.Equal(t, 0, o.value)
				if assert.NotNil(t, o.err) {
					assert.Equal(t, errors.Wrap(ErrNoResultForGivenKey, fmt.Sprintf("key(%v)", o.key)).Error(), o.err.Error())
				}

				assert.Panics(
					t, func() {
						close(o.done)
					},
					"channel should NOT be closable",
				)
			},
		},
		{
			desc: "completed outcome",
			test: func(t *testing.T) {
				o := newPendingOutcome[string, string]("abc")
				o.complete("test", assert.AnError)

				assert.Equal(t, "abc", o.key)
				assert.True(t, o.isComplete)
				assert.Equal(t, "test", o.value)
				assert.Equal(t, assert.AnError, o.err)

				// Cancel a completed outcome should have no effects
				o.cancel()

				assert.Equal(t, "abc", o.key)
				assert.True(t, o.isComplete)
				assert.Equal(t, "test", o.value)
				assert.Equal(t, assert.AnError, o.err)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}

func TestOutcome_Get(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(t *testing.T)
	}{
		{
			desc: "context was cancelled",
			test: func(t *testing.T) {
				cancelledCtx, cancelFn := context.WithCancel(context.Background())
				cancelFn()

				o := newPendingOutcome[string, int]("abc")

				result, err := o.Get(cancelledCtx)
				assert.Equal(t, 0, result)
				assert.Equal(t, context.Canceled, err)
			},
		},
		{
			desc: "outcome was completed",
			test: func(t *testing.T) {
				o := newPendingOutcome[string, int]("abc")

				go func() {
					<-time.After(200 * time.Millisecond)
					o.complete(1, assert.AnError)
				}()

				result, err := o.Get(context.Background())
				assert.Equal(t, 1, result)
				assert.Equal(t, assert.AnError, err)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}
