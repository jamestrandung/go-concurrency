// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package spread

import (
	"context"
	"time"

	"github.com/jamestrandung/go-concurrency/async"
)

// Spread evenly starts the given tasks within the specified duration. Clients can
// use the output task to block and wait for the tasks to complete if they want.
func Spread[T async.SilentTask](tasks []T, within time.Duration) async.SilentTask {
	return async.NewSilentTask(
		func(ctx context.Context) error {
			sleep := within / time.Duration(len(tasks))

			for i, t := range tasks {
				select {
				case <-ctx.Done():
					async.CancelAll(tasks[i:])
					return ctx.Err()
				default:
					t.Execute(ctx)
					time.Sleep(sleep)
				}
			}

			async.WaitAll(tasks)

			return nil
		},
	)
}
