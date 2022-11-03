package async

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueWithNoResult(
			b.Append(number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, taskCount, b.Size())
	b.Process(context.Background())

	ForkJoin(context.Background(), tasks)
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

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueWithNoResult(
			b.Append(number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	ForkJoin(context.Background(), tasks)

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

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueWithNoResult(
			b.Append(number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	ForkJoin(context.Background(), tasks)

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

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i / 2

		// All 20 tasks must still receive the right output
		tasks[i] = ContinueWithNoResult(
			b.Append(number),
			func(ctx context.Context, i int, err error) error {
				assert.Equal(t, number*10, i)
				assert.Nil(t, err)

				return nil
			},
		)
	}

	assert.Equal(t, taskCount, b.Size())
	b.Process(context.Background())

	ForkJoin(context.Background(), tasks)
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

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueWithNoResult(
			b.Append(number),
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

	ForkJoin(context.Background(), tasks)
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

	tasks := make([]SilentTask, taskCount)
	for i := 0; i < taskCount; i++ {
		number := i

		tasks[i] = ContinueWithNoResult(
			b.Append(number),
			func(ctx context.Context, i int, err error) error {
				return err
			},
		)
	}

	assert.Equal(t, 10, b.Size())

	// Shutdown should process all pending tasks
	b.Shutdown()

	ForkJoin(context.Background(), tasks)

	for i := 0; i < taskCount; i++ {
		assert.Equal(t, IsCompleted, tasks[i].State())
		assert.Equal(t, context.DeadlineExceeded, tasks[i].Error())
	}
}

func ExampleBatcher() {
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

	t1 := b.Append(1)
	t2 := b.Append(2)

	b.Process(context.Background())

	ContinueWithNoResult(
		t1, func(_ context.Context, v int, err error) error {
			fmt.Println(v)

			return nil
		},
	).ExecuteSync(context.Background())

	ContinueWithNoResult(
		t2, func(_ context.Context, v int, err error) error {
			fmt.Println(v)

			return nil
		},
	).ExecuteSync(context.Background())

	// Output:
	// [1 2]
	// 2
	// 4
}
