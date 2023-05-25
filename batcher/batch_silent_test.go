// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package batcher

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jamestrandung/go-concurrency/v2/async"

	"github.com/stretchr/testify/assert"
)

func TestSilentBatch(t *testing.T) {
	const taskCount = 10

	var wg sync.WaitGroup
	wg.Add(taskCount)

	out := make(chan int, 10)

	// Processor that multiplies items by 10 all at once
	b := NewSilentBatcher(
		func(input []interface{}) error {
			for _, number := range input {
				out <- number.(int) * 10
			}

			return nil
		},
	)

	defer b.Shutdown()

	ctx := context.Background()

	for i := 0; i < taskCount; i++ {
		number := i

		async.ContinueInSilence(
			b.Append(ctx, number), func(_ context.Context, err error) error {
				defer wg.Done()

				assert.Nil(t, err)

				return nil
			},
		).Execute(ctx)
	}

	assert.Equal(t, 10, b.Size())

	b.Process(ctx)

	wg.Wait()

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestSilentBatcher_AppendAutoProcessBySize(t *testing.T) {
	const taskCount = 10

	out := make(chan int, taskCount)

	// Processor that multiplies items by 10 all at once
	b := NewSilentBatcher(
		func(input []interface{}) error {
			for _, number := range input {
				out <- number.(int) * 10
			}

			return nil
		},
		WithAutoProcessSize(taskCount),
	)

	defer b.Shutdown()

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueInSilence(
			b.Append(ctx, number), func(_ context.Context, err error) error {
				assert.Nil(t, err)

				return err
			},
		).Execute(ctx)
	}

	async.WaitAll(tasks)

	assert.Equal(t, 0, b.Size(), "All pending tasks should have been auto processed")

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestSilentBatcher_AutoProcessOnInterval(t *testing.T) {
	const taskCount = 10

	out := make(chan int, taskCount)

	// Processor that multiplies items by 10 all at once
	b := NewSilentBatcher(
		func(input []interface{}) error {
			for _, number := range input {
				out <- number.(int) * 10
			}

			return nil
		},
		WithAutoProcessInterval(100*time.Millisecond),
	)

	defer b.Shutdown()

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueInSilence(
			b.Append(ctx, number), func(_ context.Context, err error) error {
				assert.Nil(t, err)

				return err
			},
		).Execute(ctx)
	}

	assert.Equal(t, 10, b.Size())

	async.WaitAll(tasks)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestSilentBatcher_Shutdown(t *testing.T) {
	const taskCount = 10

	out := make(chan int, taskCount)

	// Processor that multiplies items by 10 all at once
	b := NewSilentBatcher(
		func(input []interface{}) error {
			for _, number := range input {
				out <- number.(int) * 10
			}

			return nil
		},
	)

	ctx := context.Background()

	for i := 0; i < taskCount; i++ {
		number := i

		async.ContinueInSilence(
			b.Append(ctx, number), func(_ context.Context, err error) error {
				assert.Nil(t, err)

				return err
			},
		).Execute(ctx)
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, i*10, <-out)
	}
}

func TestSilentBatcher_ShutdownWithTimeout(t *testing.T) {
	const taskCount = 10

	// Processor that multiplies items by 10 all at once
	b := NewSilentBatcher(
		func(input []interface{}) error {
			time.Sleep(100 * time.Millisecond)

			return nil
		},
		WithShutdownGraceDuration(50*time.Millisecond),
	)

	ctx := context.Background()

	tasks := make([]async.SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = async.ContinueInSilence(
			b.Append(ctx, number), func(_ context.Context, err error) error {
				return err
			},
		).Execute(ctx)
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	async.WaitAll(tasks)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, async.IsCompleted, tasks[i].State())
		assert.Equal(t, context.DeadlineExceeded, tasks[i].Error())
	}
}

func ExampleSilentBatcher() {
	var wg sync.WaitGroup
	wg.Add(2)

	ctx := context.Background()

	b := NewSilentBatcher(
		func(input []interface{}) error {
			fmt.Println(input)
			return nil
		},
	)

	async.ContinueInSilence(
		b.Append(ctx, 1), func(_ context.Context, err error) error {
			wg.Done()

			return nil
		},
	).Execute(ctx)

	async.ContinueInSilence(
		b.Append(ctx, 2), func(_ context.Context, err error) error {
			wg.Done()

			return nil
		},
	).Execute(ctx)

	b.Process(ctx)

	wg.Wait()

	// Output:
	// [1 2]
}
