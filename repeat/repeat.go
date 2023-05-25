// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package repeat

import (
    "context"
    "fmt"
    "runtime/debug"
    "time"

    "github.com/jamestrandung/go-concurrency/v2/async"
)

// Repeat executes the given action asynchronously on a pre-determined interval. The repeating
// process can be stopped by cancelling the output task. All errors and panics must be handled
// inside the action if callers want the process to continue. Otherwise, the repeat will stop.
func Repeat(interval time.Duration, action async.SilentWork) async.SilentTask {
    safeAction := func(ctx context.Context) (err error) {
        defer func() {
            if r := recover(); r != nil {
                err = fmt.Errorf("panic repeating task: %v \n %s", r, debug.Stack())
            }
        }()

        return action(ctx)
    }

    return async.NewSilentTask(
        func(ctx context.Context) error {
            for {
                select {
                case <-ctx.Done():
                    return ctx.Err()

                case <-time.After(interval):
                    if err := safeAction(ctx); err != nil {
                        return err
                    }
                }
            }
        },
    )
}
