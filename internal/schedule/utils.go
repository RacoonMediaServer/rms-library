package schedule

import (
	"context"

	"go-micro.dev/v4/logger"
)

func GetRetryWrapper(l logger.Logger, fn func(logger.Logger, context.Context) error) ExecuteFn {
	return func(ctx context.Context) Result {
		if err := fn(l, ctx); err != nil {
			l.Logf(logger.ErrorLevel, "Operation failed: %s", err)
			return Result{Result: OpResultRetry}
		}
		l.Log(logger.InfoLevel, "Complete")
		return Result{Result: OpResultDone}
	}
}
