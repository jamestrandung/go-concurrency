// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package spread

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jamestrandung/go-concurrency/v2/async"

	"github.com/stretchr/testify/assert"
)

func newTasks() []async.Task[int] {
	work := func(context.Context) (int, error) {
		return 1, nil
	}

	return async.NewTasks(work, work, work, work, work)
}

func TestSpread(t *testing.T) {
	tasks := newTasks()
	within := 200 * time.Millisecond

	// Spread and calculate the duration
	t0 := time.Now()

	spreadTask := Spread(tasks, within)
	spreadTask.ExecuteSync(context.Background())

	// Make sure we completed within duration
	dt := int(time.Now().Sub(t0).Seconds() * 1000)
	assert.True(t, dt > 150 && dt < 300, fmt.Sprintf("%v ms.", dt))

	// Make sure all tasks are done
	for _, task := range tasks {
		v, _ := task.Outcome()
		assert.Equal(t, 1, v)
	}
}

func TestSpread_Cancel(t *testing.T) {
	tasks := newTasks()
	within := 200 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	spreadTask := Spread(tasks, within)
	spreadTask.Execute(ctx)

	async.WaitAll(tasks)

	cancelled := 0
	for _, t := range tasks {
		if t.State() == async.IsCancelled {
			cancelled++
		}
	}

	assert.Equal(t, 5, cancelled)
}

func ExampleSpread() {
	tasks := newTasks()
	within := 200 * time.Millisecond

	// Spread
	spreadTask := Spread(tasks, within)
	spreadTask.ExecuteSync(context.Background())

	// Make sure all tasks are done
	for _, task := range tasks {
		v, _ := task.Outcome()
		fmt.Println(v)
	}

	// Output:
	// 1
	// 1
	// 1
	// 1
	// 1
}
