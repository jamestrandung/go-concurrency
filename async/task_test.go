// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompleted(t *testing.T) {
	task := Completed(1, assert.AnError)

	v, err := task.Outcome()
	assert.Equal(t, 1, v)
	assert.Equal(t, assert.AnError, err)
	assert.Equal(t, IsCompleted, task.State())
}

func TestNewTasks(t *testing.T) {
	work := func(context.Context) (interface{}, error) {
		return 1, nil
	}

	tasks := NewTasks(work, work, work)
	assert.Equal(t, 3, len(tasks))
}

func TestOutcome(t *testing.T) {
	task := Invoke(
		context.Background(), func(context.Context) (interface{}, error) {
			return 1, nil
		},
	)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			o, _ := task.Outcome()
			wg.Done()
			assert.Equal(t, 1, o.(int))
		}()
	}
	wg.Wait()
}

func TestOutcomeTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	task := Invoke(
		ctx, func(context.Context) (interface{}, error) {
			time.Sleep(500 * time.Millisecond)
			return 1, nil
		},
	)

	_, err := task.Outcome()
	assert.Equal(t, "context deadline exceeded", err.Error())
}

func TestResultOrDefault(t *testing.T) {
	task := Invoke(
		context.Background(), func(context.Context) (interface{}, error) {
			return 1, assert.AnError
		},
	)

	var wg sync.WaitGroup
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			o := task.ResultOrDefault(0)
			wg.Done()
			assert.Equal(t, 0, o.(int))
		}()
	}
	wg.Wait()
}

func TestContinueWithChain(t *testing.T) {
	task1 := NewTask(
		func(context.Context) (int, error) {
			return 1, nil
		},
	)

	task2 := ContinueWith(
		task1, func(_ context.Context, result int, _ error) (int, error) {
			return result + 1, nil
		},
	)

	task3 := ContinueWith(
		task2, func(_ context.Context, result int, _ error) (int, error) {
			return result + 1, nil
		},
	)

	result, err := task3.Run(context.Background()).Outcome()
	assert.Equal(t, 3, result)
	assert.Nil(t, err)
}

func TestContinueInSilenceChain(t *testing.T) {
	num := 0

	task1 := NewSilentTask(
		func(context.Context) error {
			num += 1
			return nil
		},
	)

	task2 := ContinueInSilence(
		task1, func(context.Context, error) error {
			num += 1
			return nil
		},
	)

	task3 := ContinueInSilence(
		task2, func(context.Context, error) error {
			num += 1
			return nil
		},
	)

	task3.Execute(context.Background()).Wait()
	assert.Equal(t, 3, num)
	assert.Nil(t, task3.Error())
}

func TestContinueWithNoResultChain(t *testing.T) {
	num := 0

	task1 := NewTask(
		func(context.Context) (int, error) {
			return 1, nil
		},
	)

	task2 := ContinueWithNoResult(
		task1, func(_ context.Context, result int, _ error) error {
			num = result + 1
			return nil
		},
	)

	task3 := ContinueInSilence(
		task2, func(context.Context, error) error {
			num += 1
			return nil
		},
	)

	task3.Execute(context.Background()).Wait()
	assert.Equal(t, 3, num)
	assert.Nil(t, task3.Error())
}

func TestContinueWithResultChain(t *testing.T) {
	num := 0

	task1 := NewTask(
		func(context.Context) (int, error) {
			return 1, nil
		},
	)

	task2 := ContinueWithNoResult(
		task1, func(_ context.Context, result int, _ error) error {
			num = result + 1
			return nil
		},
	)

	task3 := ContinueWithResult(
		task2, func(context.Context, error) (int, error) {
			return num + 1, nil
		},
	)

	result, err := task3.Run(context.Background()).Outcome()
	assert.Equal(t, 3, result)
	assert.Nil(t, err)
}

func TestContinueTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	first := Invoke(
		ctx, func(context.Context) (int, error) {
			return 5, nil
		},
	)

	second := ContinueWith(
		first, func(_ context.Context, result int, err error) (int, error) {
			time.Sleep(500 * time.Millisecond)
			return result, err
		},
	).Run(ctx)

	r1, err1 := first.Outcome()
	assert.Equal(t, 5, r1)
	assert.Nil(t, err1)

	_, err2 := second.Outcome()
	assert.Equal(t, "context deadline exceeded", err2.Error())
}

func TestTaskCancelStarted(t *testing.T) {
	task := Invoke(
		context.Background(), func(context.Context) (interface{}, error) {
			time.Sleep(500 * time.Millisecond)
			return 1, nil
		},
	)

	task.Cancel()

	_, err := task.Outcome()
	assert.Equal(t, fmt.Errorf("task cancelled with reason: %s", ErrDefaultCancelReason.Error()), err)
}

func TestTaskCancelRunning(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	task := Invoke(
		context.Background(), func(context.Context) (interface{}, error) {
			defer wg.Done()

			time.Sleep(500 * time.Millisecond)
			return 1, nil
		},
	)

	time.Sleep(10 * time.Millisecond)

	task.Cancel()

	wg.Wait()

	_, err := task.Outcome()
	assert.Equal(t, fmt.Errorf("task cancelled with reason: %s", ErrDefaultCancelReason.Error()), err)
}

func TestTaskCancelTwice(t *testing.T) {
	task := Invoke(
		context.Background(), func(context.Context) (interface{}, error) {
			time.Sleep(500 * time.Millisecond)
			return 1, nil
		},
	)

	assert.NotPanics(
		t, func() {
			for i := 0; i < 100; i++ {
				task.Cancel()
			}
		},
	)

	_, err := task.Outcome()
	assert.Equal(t, fmt.Errorf("task cancelled with reason: %s", ErrDefaultCancelReason.Error()), err)
}
