// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
	"context"
	"golang.org/x/sync/errgroup"
)

// ForkJoinFailFast executes given tasks in parallel and waits for the 1st task to fail and
// returns immediately or for ALL to complete successfully before returning.
//
// Note: task cannot be nil
func ForkJoinFailFast[T SilentTask](ctx context.Context, tasks []T) error {
	g, groupCtx := errgroup.WithContext(ctx)
	for _, t := range tasks {
		t := t
		g.Go(
			func() error {
				return t.ExecuteSync(groupCtx).Error()
			},
		)
	}

	return g.Wait()
}

// ForkJoin executes given tasks in parallel and waits for ALL to complete before returning.
//
// Note: task cannot be nil
func ForkJoin[T SilentTask](ctx context.Context, tasks []T) {
	for _, t := range tasks {
		t.Execute(ctx)
	}

	WaitAll(tasks)
}

// WaitAll waits for all executed tasks to finish.
//
// Note: task cannot be nil
func WaitAll[T SilentTask](tasks []T) {
	for _, t := range tasks {
		t.Wait()
	}
}

// CancelAll cancels all given tasks.
//
// Note: task cannot be nil
func CancelAll[T SilentTask](tasks []T) {
	for _, t := range tasks {
		t.Cancel()
	}
}
