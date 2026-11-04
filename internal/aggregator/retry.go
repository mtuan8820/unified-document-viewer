package aggregator

import (
	"context"
	"time"
)

type Effector[T any] func(ctx context.Context) (T, error)

func Retry[T any](effector Effector[T], retries int, delay time.Duration) Effector[T] {
	return func(ctx context.Context) (T, error) {
		var zero T
		for attempt := 0; ; attempt++ {
			if err := ctx.Err(); err != nil {
				return zero, err
			}

			response, err := effector(ctx)

			if err == nil || attempt >= retries {
				return response, err
			}

			select {
			case <-ctx.Done():
				//reach timeout
				return zero, ctx.Err()
			case <-time.After(delay):
				// delay elapsed
			}
		}
	}
}
