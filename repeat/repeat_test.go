// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package repeat

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRepeat(t *testing.T) {
	assert.NotPanics(
		t, func() {
			out := make(chan bool, 1)
			task := Repeat(
				time.Millisecond*100, func(context.Context) error {
					defer func() {
						recover()
					}()

					out <- true
					return nil
				},
			)

			task.Execute(context.Background())

			<-out
			v := <-out

			assert.True(t, v)

			task.Cancel()
			close(out)
		},
	)
}

func ExampleRepeat() {
	out := make(chan bool, 1)
	task := Repeat(
		time.Nanosecond*10, func(context.Context) error {
			defer func() {
				recover()
			}()

			out <- true
			return nil
		},
	)

	task.Execute(context.Background())

	<-out
	v := <-out

	fmt.Println(v)

	task.Cancel()
	close(out)

	// Output:
	// true
}

func TestRepeatPanic(t *testing.T) {
	assert.NotPanics(
		t, func() {
			var counter int32
			task := Repeat(
				time.Nanosecond*10, func(context.Context) error {
					atomic.AddInt32(&counter, 1)
					panic("test")
				},
			)

			task.Execute(context.Background())

			err := task.Error()
			assert.NotNil(t, err)
		},
	)
}
