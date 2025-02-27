// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package jitter

import (
	"context"
	"fmt"

	"github.com/jamestrandung/go-concurrency/v2/async"
)

func ExampleDoJitter() {
	t1 := async.InvokeInSilence(
		context.Background(), func(ctx context.Context) error {
			DoJitter(
				func() {
					fmt.Println("do something after random jitter")
				}, 1000,
			)

			return nil
		},
	)

	t2 := AddJitterT(
		async.NewTask(
			func(ctx context.Context) (int, error) {
				fmt.Println("return 1 after random jitter")
				return 1, nil
			},
		),
		1000,
	).Run(context.Background())

	t3 := AddJitterST(
		async.NewSilentTask(
			func(ctx context.Context) error {
				fmt.Println("return nil after random jitter")
				return nil
			},
		),
		1000,
	).Execute(context.Background())

	async.WaitAll([]async.SilentTask{t1, t2, t3})
}
