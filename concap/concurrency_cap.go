// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package concap

import (
	"context"
	"runtime"

	"github.com/jamestrandung/go-concurrency/async"
)

func cancelRemainingTasks[T async.SilentTask](tasks <-chan T) {
	for {
		select {
		case t, ok := <-tasks:
			if ok {
				t.Cancel()
			}
		default:
			return
		}
	}
}

// RunWithConcurrencyLevelC runs the given tasks up to the max concurrency level. Clients can use the output
// task to block and wait for the tasks to complete if they want.
//
// Note: When `ctx` is cancelled, we spawn a new goroutine to cancel all remaining tasks in the given channel.
// To avoid memory leak, clients MUST make sure new tasks will eventually stop arriving once `ctx` is cancelled
// so that the new goroutine can return.
func RunWithConcurrencyLevelC[T async.SilentTask](concurrencyLevel int, tasks <-chan T) async.SilentTask {
	if concurrencyLevel <= 0 {
		concurrencyLevel = runtime.NumCPU()
	}

	return async.NewSilentTask(
		func(ctx context.Context) error {
			workers := make(chan int, concurrencyLevel)
			concurrentTasks := make([]async.SilentTask, concurrencyLevel)

			// Generate worker IDs
			for id := 0; id < concurrencyLevel; id++ {
				workers <- id
			}

			for {
				select {
				// Context cancelled
				case <-ctx.Done():
					go cancelRemainingTasks(tasks)
					waitForNonNilTasks(concurrentTasks)
					return ctx.Err()

				// Worker available
				case workerID := <-workers:
					select {
					// Worker is waiting for job when context is cancelled
					case <-ctx.Done():
						go cancelRemainingTasks(tasks)
						waitForNonNilTasks(concurrentTasks)
						return ctx.Err()

					case t, ok := <-tasks:
						// Task channel is closed
						if !ok {
							waitForNonNilTasks(concurrentTasks)
							return nil
						}

						concurrentTasks[workerID] = t

						// Return the worker to the common pool
						async.ContinueInSilence(
							t, func(context.Context, error) error {
								workers <- workerID
								return nil
							},
						).Execute(ctx)
					}
				}
			}
		},
	)
}

func waitForNonNilTasks(tasks []async.SilentTask) {
	nonNilTasks := make([]async.SilentTask, 0, len(tasks))
	for _, task := range tasks {
		if task != nil {
			nonNilTasks = append(nonNilTasks, task)
		}
	}

	async.WaitAll(nonNilTasks)
}

// RunWithConcurrencyLevelS runs the given tasks up to the max concurrency level. Clients
// can use the output task to block and wait for the tasks to complete if they want.
func RunWithConcurrencyLevelS[T async.SilentTask](concurrencyLevel int, tasks []T) async.SilentTask {
	if concurrencyLevel == 0 {
		concurrencyLevel = runtime.NumCPU()
	}

	return async.NewSilentTask(
		func(ctx context.Context) error {
			sem := make(chan struct{}, concurrencyLevel)

			for i, t := range tasks {
				select {
				case <-ctx.Done():
					async.CancelAll(tasks[i:])
					return ctx.Err()
				case sem <- struct{}{}:
					// Return the worker to the common pool
					async.ContinueInSilence(
						t, func(context.Context, error) error {
							<-sem
							return nil
						},
					).Execute(ctx)
				}
			}

			async.WaitAll(tasks)

			return nil
		},
	)
}
