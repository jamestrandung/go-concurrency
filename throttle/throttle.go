// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package throttle

import (
	"context"
	"time"

	"github.com/jamestrandung/go-concurrency/v2/async"

	"golang.org/x/time/rate"
)

// Throttle starts the given tasks at the specified rate. Clients can use
// the output task to block and wait for the tasks to complete if they want.
func Throttle[T async.SilentTask](tasks []T, rateLimit int, every time.Duration) async.SilentTask {
	return async.NewSilentTask(
		func(ctx context.Context) error {
			limiter := rate.NewLimiter(rate.Every(every/time.Duration(rateLimit)), 1)

			for i, t := range tasks {
				select {
				case <-ctx.Done():
					async.CancelAll(tasks[i:])
					return ctx.Err()
				default:
					if err := limiter.Wait(ctx); err == nil {
						t.Execute(ctx)
					}
				}
			}

			async.WaitAll(tasks)

			return nil
		},
	)
}
