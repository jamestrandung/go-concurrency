// Copyright (c) 2022 James Tran Dung, All rights reserved.
// Use of this source code is governed by an MIT-style license that can be found in the LICENSE file

package async

import (
    "context"
    "errors"
    "fmt"
    "runtime/debug"
    "sync"
    "sync/atomic"
    "time"
)

// ErrDefaultCancelReason is default reason when none is provided.
var ErrDefaultCancelReason = errors.New("no reason provided")

var now = time.Now

// SilentWork represents a unit of work to execute in
// silence like background works that return no values.
type SilentWork func(context.Context) error

// Work represents a unit of work to execute that is
// expected to return a value of a particular type.
type Work[T any] func(context.Context) (T, error)

// PanicRecoverWork represents a unit of work to be
// executed when a panic occurs.
type PanicRecoverWork func(any)

// State represents the state enumeration for a task.
type State byte

// Various task states.
const (
    IsCreated   State = iota // IsCreated represents a newly created task
    IsRunning                // IsRunning represents a task which is currently running
    IsCompleted              // IsCompleted represents a task which was completed successfully or errored out
    IsCancelled              // IsCancelled represents a task which was cancelled or has timed out
)

type signal chan struct{}

// SilentTask represents a unit of work to complete in silence
// like background works that return no values.
//go:generate mockery --name SilentTask --case underscore --inpackage
type SilentTask interface {
    // WithRecoverAction attaches the given recover action with task so that
    // it can be executed when a panic occurs.
    WithRecoverAction(recoverAction PanicRecoverWork)
    // Execute starts this task asynchronously.
    Execute(ctx context.Context) SilentTask
    // ExecuteSync starts this task synchronously.
    ExecuteSync(ctx context.Context) SilentTask
    // Wait waits for this task to complete.
    Wait()
    // Cancel changes the state of this task to `Cancelled`.
    Cancel()
    // CancelWithReason changes the state of this task to `Cancelled` with the given reason.
    CancelWithReason(error)
    // Error returns the error that occurred when this task was executed.
    Error() error
    // State returns the current state of this task. This operation is non-blocking.
    State() State
    // Duration returns the duration of this task.
    Duration() time.Duration
}

// Task represents a unit of work that is expected to return
// a value of a particular type.
//go:generate mockery --name Task --case underscore --inpackage
type Task[T any] interface {
    SilentTask
    // Run starts this task asynchronously.
    Run(ctx context.Context) Task[T]
    // RunSync starts this task synchronously.
    RunSync(ctx context.Context) Task[T]
    // Outcome waits for this task to complete and returns the final result & error.
    Outcome() (T, error)
    // ResultOrDefault waits for this task to complete and returns the final result if
    // there's no error or the default result if there's an error.
    ResultOrDefault(T) T
}

type outcome[T any] struct {
    result T
    err    error
}

type task[T any] struct {
    state      int32      // The current async.State of this task
    cancel     chan error // The channel for cancelling this task
    cancelOnce sync.Once
    done       signal           // The channel for indicating this task has completed
    action     Work[T]          // The work to do
    outcome    outcome[T]       // This is used to store the outcome of this task
    rAction    PanicRecoverWork // The work to do when a panic occurs
    duration   time.Duration    // The duration of this task, in nanoseconds
}

// Completed returns a completed task with the given result and error.
func Completed[T any](result T, err error) Task[T] {
    done := make(signal)
    close(done)

    return &task[T]{
        state: int32(IsCompleted),
        done:  done,
        outcome: outcome[T]{
            result: result,
            err:    err,
        },
        duration: time.Duration(0),
    }
}

// NewTask creates a new Task.
func NewTask[T any](action Work[T]) Task[T] {
    return &task[T]{
        action: action,
        done:   make(signal, 1),
        cancel: make(chan error, 1),
    }
}

// NewTasks creates a group of new Task.
func NewTasks[T any](actions ...Work[T]) []Task[T] {
    tasks := make([]Task[T], 0, len(actions))

    for _, action := range actions {
        tasks = append(tasks, NewTask(action))
    }

    return tasks
}

// NewSilentTask creates a new SilentTask.
func NewSilentTask(action SilentWork) SilentTask {
    return &task[struct{}]{
        action: func(taskCtx context.Context) (struct{}, error) {
            return struct{}{}, action(taskCtx)
        },
        done:   make(signal, 1),
        cancel: make(chan error, 1),
    }
}

// NewSilentTasks creates a group of new SilentTask.
func NewSilentTasks(actions ...SilentWork) []SilentTask {
    tasks := make([]SilentTask, 0, len(actions))

    for _, action := range actions {
        tasks = append(tasks, NewSilentTask(action))
    }

    return tasks
}

// Invoke creates a new Task and runs it asynchronously.
func Invoke[T any](ctx context.Context, action Work[T]) Task[T] {
    return NewTask(action).Run(ctx)
}

// InvokeInSilence creates a new SilentTask and runs it asynchronously.
func InvokeInSilence(ctx context.Context, action SilentWork) SilentTask {
    return Invoke(
        ctx, func(taskCtx context.Context) (struct{}, error) {
            return struct{}{}, action(taskCtx)
        },
    )
}

// ContinueWith proceeds with the next task once the current one is finished. When we have a chain of tasks like
// A -> B -> C, executing C will trigger A & B as well. However, executing A will NOT trigger B & C.
func ContinueWith[T any, S any](currentTask Task[T], nextAction func(context.Context, T, error) (S, error)) Task[S] {
    return NewTask(
        func(taskCtx context.Context) (S, error) {
            // Attempt to run the current task in case it has not been executed.
            currentTask.Execute(taskCtx)

            result, err := currentTask.Outcome()
            return nextAction(taskCtx, result, err)
        },
    )
}

// ContinueWithNoResult proceeds with the next task once the current one is finished. When we have a chain of tasks
// like A -> B -> C, executing C will trigger A & B as well. However, executing A will NOT trigger B & C.
func ContinueWithNoResult[T any](currentTask Task[T], nextAction func(context.Context, T, error) error) SilentTask {
    return NewSilentTask(
        func(taskCtx context.Context) error {
            // Attempt to run the current task in case it has not been executed.
            currentTask.Execute(taskCtx)

            result, err := currentTask.Outcome()
            return nextAction(taskCtx, result, err)
        },
    )
}

// ContinueInSilence proceeds with the next task once the current one is finished. When we have a chain of tasks like
// A -> B -> C, executing C will trigger A & B as well. However, executing A will NOT trigger B & C.
func ContinueInSilence(currentTask SilentTask, nextAction func(context.Context, error) error) SilentTask {
    return NewSilentTask(
        func(taskCtx context.Context) error {
            // Attempt to run the current task in case it has not been executed.
            currentTask.Execute(taskCtx)

            return nextAction(taskCtx, currentTask.Error())
        },
    )
}

// ContinueWithResult proceeds with the next task once the current one is finished. When we have a chain of tasks like
// A -> B -> C, executing C will trigger A & B as well. However, executing A will NOT trigger B & C.
func ContinueWithResult[T any](currentTask SilentTask, nextAction func(context.Context, error) (T, error)) Task[T] {
    return NewTask(
        func(taskCtx context.Context) (T, error) {
            // Attempt to run the current task in case it has not been executed.
            currentTask.Execute(taskCtx)

            return nextAction(taskCtx, currentTask.Error())
        },
    )
}

func (t *task[T]) WithRecoverAction(recoverAction PanicRecoverWork) {
    t.rAction = recoverAction
}

func (t *task[T]) Outcome() (T, error) {
    <-t.done
    return t.outcome.result, t.outcome.err
}

func (t *task[T]) ResultOrDefault(defaultResult T) T {
    <-t.done

    if t.outcome.err != nil {
        return defaultResult
    }

    return t.outcome.result
}

func (t *task[T]) Error() error {
    <-t.done
    return t.outcome.err
}

func (t *task[T]) Wait() {
    <-t.done
}

func (t *task[T]) State() State {
    v := atomic.LoadInt32(&t.state)
    return State(v)
}

func (t *task[T]) Duration() time.Duration {
    return t.duration
}

func (t *task[T]) Run(ctx context.Context) Task[T] {
    go t.doRun(ctx)
    return t
}

func (t *task[T]) RunSync(ctx context.Context) Task[T] {
    ok := t.doRun(ctx)
    if !ok {
        // Task already executed
        t.Wait()
    }

    return t
}

func (t *task[T]) Execute(ctx context.Context) SilentTask {
    go t.doRun(ctx)
    return t
}

func (t *task[T]) ExecuteSync(ctx context.Context) SilentTask {
    ok := t.doRun(ctx)
    if !ok {
        // Task already executed
        t.Wait()
    }

    return t
}

func (t *task[T]) Cancel() {
    t.CancelWithReason(ErrDefaultCancelReason)
}

func (t *task[T]) CancelWithReason(err error) {
    if err == nil {
        err = ErrDefaultCancelReason
    }

    // If the task was created but never started, transition directly to cancelled state
    // and close the done channel and set the error.
    if t.changeState(IsCreated, IsCancelled) {
        t.outcome = outcome[T]{err: fmt.Errorf("task cancelled with reason: %s", err.Error())}
        close(t.done)
        return
    }

    // Attempt to cancel the task if it's in the running state
    if t.cancel == nil {
        return
    }

    t.cancelOnce.Do(
        func() {
            t.cancel <- err
            close(t.cancel)
        },
    )
}

func (t *task[T]) doRun(ctx context.Context) bool {
    // Prevent from running the same task twice
    if !t.changeState(IsCreated, IsRunning) {
        return false
    }

    // When this task get cancelled, this `ctx` will get cancelled as well
    ctxWithCancel, cancel := context.WithCancel(ctx)
    defer cancel()

    // Notify everyone of the completion/error state
    defer close(t.done)

    // Execute the task
    startedAt := now().UnixNano()
    outcomeCh := make(chan outcome[T], 1)
    go func() {
        defer close(outcomeCh)
        // Convert panics into standard errors for clients to handle gracefully
        defer func() {
            if r := recover(); r != nil {
                // Perform the registered recover action
                t.executeRecoverAction(r)

                outcomeCh <- outcome[T]{err: fmt.Errorf("panic executing async task: %v \n %s", r, debug.Stack())}
            }
        }()

        r, e := t.action(ctxWithCancel)
        outcomeCh <- outcome[T]{result: r, err: e}
    }()

    select {
    // In case of a manual task cancellation, set the outcome and transition
    // to the cancelled state.
    case err := <-t.cancel:
        if err == nil {
            err = ErrDefaultCancelReason
        }

        t.duration = time.Nanosecond * time.Duration(now().UnixNano()-startedAt)
        t.outcome = outcome[T]{err: fmt.Errorf("task cancelled with reason: %s", err.Error())}
        t.changeState(IsRunning, IsCancelled)

    // In case of the context timeout or other error, change the state of the
    // task to cancelled and return right away.
    case <-ctxWithCancel.Done():
        t.duration = time.Nanosecond * time.Duration(now().UnixNano()-startedAt)
        t.outcome = outcome[T]{err: ctxWithCancel.Err()}
        t.changeState(IsRunning, IsCancelled)

    // In case where we got an outcome (happy path).
    case o := <-outcomeCh:
        t.duration = time.Nanosecond * time.Duration(now().UnixNano()-startedAt)
        t.outcome = o
        t.changeState(IsRunning, IsCompleted)
    }

    return true
}

func (t *task[T]) executeRecoverAction(recoverDetails any) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("panic executing recover action: %v \n %s \n", r, debug.Stack())
        }
    }()

    if t.rAction != nil {
        t.rAction(recoverDetails)
    }
}

func (t *task[T]) changeState(from, to State) bool {
    return atomic.CompareAndSwapInt32(&t.state, int32(from), int32(to))
}
