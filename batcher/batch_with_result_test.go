// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package batcher

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jamestrandung/go-concurrency/async"
	"github.com/stretchr/testify/assert"
)

func TestBatch(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) ([]int, error) {
			result := make([]int, len(input))
			for idx, number := range input {
				result[idx] = number * 10
			}

			return result, nil
		},
		nil,
	)

	defer b.Shutdown()

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueWithNoResult(
			b.Append(ctx, number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, taskCount, b.Size())
	b.Process(ctx)

	async.ForkJoin(ctx, tasks)
}

func TestBatcher_AppendAutoProcessBySize(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) ([]int, error) {
			result := make([]int, len(input))
			for idx, number := range input {
				result[idx] = number * 10
			}

			return result, nil
		},
		nil,
		WithAutoProcessSize(taskCount),
	)

	defer b.Shutdown()

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueWithNoResult(
			b.Append(ctx, number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	async.ForkJoin(ctx, tasks)

	assert.Equal(t, 0, b.Size(), "All pending tasks should have been auto processed")
}

func TestBatcher_AutoProcessOnInterval(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) ([]int, error) {
			result := make([]int, len(input))
			for idx, number := range input {
				result[idx] = number * 10
			}

			return result, nil
		},
		nil,
		WithAutoProcessInterval(100*time.Millisecond),
	)

	defer b.Shutdown()

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueWithNoResult(
			b.Append(ctx, number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	async.ForkJoin(ctx, tasks)

	assert.Equal(t, 0, b.Size(), "All pending tasks should have been auto processed")
}

func TestBatcher_SingleFlightEnabled(t *testing.T) {
	const taskCount = 20

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) ([]int, error) {
			assert.Equal(t, taskCount/2, len(input), "We must not receive duplicate inputs")

			result := make([]int, len(input))
			for idx, number := range input {
				result[idx] = number * 10
			}

			return result, nil
		},
		func(i int) any {
			return i
		},
	)

	defer b.Shutdown()

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i / 2

		// All 20 tasks must still receive the right output
		tasks[i] = async.ContinueWithNoResult(
			b.Append(ctx, number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, taskCount, b.Size())
	b.Process(ctx)

	async.ForkJoin(ctx, tasks)
}

func TestBatcher_Shutdown(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) ([]int, error) {
			result := make([]int, len(input))
			for idx, number := range input {
				result[idx] = number * 10
			}

			return result, nil
		},
		nil,
	)

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueWithNoResult(
			b.Append(ctx, number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	async.ForkJoin(ctx, tasks)
}

func TestBatcher_ShutdownWithTimeout(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewBatcher(
		func(input []int) ([]int, error) {
			time.Sleep(100 * time.Millisecond)

			return nil, nil
		},
		nil,
		WithShutdownGraceDuration(50*time.Millisecond),
	)

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueWithNoResult(
			b.Append(ctx, number),
			func(ctx context.Context, i int, err error) error {
				return err
			},
		)
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	async.ForkJoin(ctx, tasks)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, async.IsCompleted, tasks[i].State())
		assert.Equal(t, context.DeadlineExceeded, tasks[i].Error())
	}
}

func ExampleBatcher() {
	ctx := context.Background()

	b := NewBatcher(
		func(input []int) ([]int, error) {
			fmt.Println(input)

			result := make([]int, len(input))
			for idx, number := range input {
				result[idx] = number * 2
			}

			return result, nil
		},
		nil,
	)

	defer b.Shutdown()

	t1 := b.Append(ctx, 1)
	t2 := b.Append(ctx, 2)

	b.Process(ctx)

	async.ContinueWithNoResult(
		t1, func(_ context.Context, v int, err error) error {
			fmt.Println(v)

			return nil
		},
	).ExecuteSync(ctx)

	async.ContinueWithNoResult(
		t2, func(_ context.Context, v int, err error) error {
			fmt.Println(v)

			return nil
		},
	).ExecuteSync(ctx)

	// Output:
	// [1 2]
	// 2
	// 4
}
