// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package jitter

import (
	"context"
	"math/rand"
	"time"

	"github.com/jamestrandung/go-concurrency/v2/async"
)

// DoJitter adds a random jitter before executing doFn, then returns the jitter duration.
func DoJitter(doFn func(), maxJitterDurationInMilliseconds int) int {
	randomJitterDuration := waitForRandomJitter(maxJitterDurationInMilliseconds)

	doFn()

	return randomJitterDuration
}

// AddJitterT adds a random jitter before executing the given Task.
func AddJitterT[T any](t async.Task[T], maxJitterDurationInMilliseconds int) async.Task[T] {
	return async.NewTask(
		func(ctx context.Context) (T, error) {
			waitForRandomJitter(maxJitterDurationInMilliseconds)

			t.ExecuteSync(ctx)

			return t.Outcome()
		},
	)
}

// AddJitterST adds a random jitter before executing the given SilentTask.
func AddJitterST(t async.SilentTask, maxJitterDurationInMilliseconds int) async.SilentTask {
	return async.NewSilentTask(
		func(ctx context.Context) error {
			waitForRandomJitter(maxJitterDurationInMilliseconds)

			t.ExecuteSync(ctx)

			return t.Error()
		},
	)
}

func waitForRandomJitter(maxJitterDurationInMilliseconds int) int {
	rand.Seed(time.Now().UnixNano())
	min := 0
	max := maxJitterDurationInMilliseconds

	randomJitterDuration := rand.Intn(max-min+1) + min

	<-time.After(time.Duration(randomJitterDuration) * time.Millisecond)

	return randomJitterDuration
}
