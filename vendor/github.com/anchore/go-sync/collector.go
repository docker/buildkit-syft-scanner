package sync

import (
	"context"
	"errors"
	"iter"
	"sync"
)

// Collect iterates over the provided iterator, executing the processor in parallel to map each incoming value to a result.
// The accumulator is used to apply the results, with an exclusive lock; accumulator will never execute in parallel.
// All errors returned from processor functions will be joined with errors.Join as the returned error.
func Collect[From, To any](ctx *context.Context, executorName string, iterator iter.Seq[From], processor func(From) (To, error), accumulator func(From, To)) error {
	if processor == nil {
		panic("no processor provided to Collect")
	}
	if ctx == nil || *ctx == nil {
		ctx = emptyContextPtr
	}
	var errs []error
	var lock sync.Mutex
	var wg sync.WaitGroup
	executor := ContextExecutor(ctx, executorName)
	for i := range iterator {
		// skip queuing any more values
		if (*ctx).Err() != nil {
			break
		}
		wg.Add(1)
		executor.Go(func() {
			defer wg.Done()
			// we may have queued many functions when canceled
			if (*ctx).Err() != nil {
				return
			}
			result, err := processor(i)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				errs = append(errs, err)
			}
			if accumulator != nil {
				accumulator(i, result)
			}
		})
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-(*ctx).Done():
	case <-done:
	}

	return errors.Join(errs...)
}

// CollectSlice is a specialized Collect call which appends results to a slice
func CollectSlice[From, To any](ctx *context.Context, executorName string, values iter.Seq[From], processor func(From) (To, error), slice *[]To) error {
	return Collect(ctx, executorName, values, processor, func(_ From, value To) {
		*slice = append(*slice, value)
	})
}

// CollectMap is a specialized Collect call which fills a map using the incoming value as a key, mapped to the result
func CollectMap[From comparable, To any](ctx *context.Context, executorName string, values iter.Seq[From], processor func(From) (To, error), result map[From]To) error {
	return Collect(ctx, executorName, values, processor, func(key From, value To) {
		result[key] = value
	})
}

// ToSeq converts a slice to an iter.Seq
func ToSeq[T any](values []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range values {
			if !yield(value) {
				return
			}
		}
	}
}

// ToSlice takes an iter.Seq and returns a slice of the values returned
func ToSlice[T any](values iter.Seq[T]) (everything []T) {
	for v := range values {
		everything = append(everything, v)
	}
	return everything
}
