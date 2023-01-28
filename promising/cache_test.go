package promising

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestCache_Destroy(t *testing.T) {
	c := &cache[string, int]{
		outcomes: make(map[string]IOutcome[int]),
	}

	assert.False(t, c.isDestroyed)

	mockOutcome1 := &MockIOutcome[int]{}
	mockOutcome1.On("cancel").Once()

	c.outcomes["abc"] = mockOutcome1

	mockOutcome2 := &MockIOutcome[int]{}
	mockOutcome2.On("cancel").Once()

	c.outcomes["def"] = mockOutcome2

	c.Destroy()

	assert.True(t, c.isDestroyed)
	mockOutcome1.AssertExpectations(t)
	mockOutcome2.AssertExpectations(t)
}

func TestCache_GetOutcome(t *testing.T) {
	scenarios := []struct {
		desc string
		test func(t *testing.T)
	}{
		{
			desc: "cache was destroyed, outcome exists",
			test: func(t *testing.T) {
				c := &cache[string, int]{
					outcomes: make(map[string]IOutcome[int]),
				}

				key := "abc"

				mockOutcome := &MockIOutcome[int]{}
				mockOutcome.On("cancel").Once()

				c.outcomes[key] = mockOutcome

				c.Destroy()

				o := c.GetOutcome(key)
				assert.Equal(t, mockOutcome, o)
			},
		},
		{
			desc: "cache was destroyed, outcome doesn't exist",
			test: func(t *testing.T) {
				c := NewCache[string, int]()
				c.Destroy()

				key := "abc"
				o := c.GetOutcome(key).(*outcome[string, int])

				ctxWithTimeout, cancelFn := context.WithTimeout(context.Background(), 200*time.Millisecond)
				defer cancelFn()

				// Should return immediately with an error. Otherwise, context will time out.
				val, err := o.Get(ctxWithTimeout)
				assert.Equal(t, 0, val)
				assert.Equal(t, errors.Wrap(ErrNoMoreResultIncoming, fmt.Sprintf("key: %v", key)).Error(), err.Error())
			},
		},
		{
			desc: "outcome exists",
			test: func(t *testing.T) {
				c := &cache[string, int]{
					outcomes: make(map[string]IOutcome[int]),
				}

				key := "abc"

				mockOutcome := &MockIOutcome[int]{}
				mockOutcome.On("cancel").Once()

				c.outcomes[key] = mockOutcome

				o := c.GetOutcome(key)
				assert.Equal(t, mockOutcome, o)
			},
		},
		{
			desc: "outcome doesn't exist",
			test: func(t *testing.T) {
				c := &cache[string, int]{
					outcomes: make(map[string]IOutcome[int]),
				}

				key := "abc"

				o := c.GetOutcome(key).(*outcome[string, int])
				assert.Equal(t, o, c.outcomes[key])
				assert.Equal(t, key, o.key)
				assert.False(t, o.isComplete)
				assert.Equal(t, 0, o.value)
				assert.Nil(t, o.err)
			},
		},
	}

	for _, scenario := range scenarios {
		sc := scenario

		t.Run(sc.desc, sc.test)
	}
}

func TestCache_Complete(t *testing.T) {
	c := &cache[string, int]{
		outcomes: make(map[string]IOutcome[int]),
	}

	mockOutcome := &MockIOutcome[int]{}
	mockOutcome.On("complete", 1, assert.AnError).Once()

	c.outcomes["abc"] = mockOutcome

	c.Complete("abc", 1, assert.AnError)

	mockOutcome.AssertExpectations(t)
}
