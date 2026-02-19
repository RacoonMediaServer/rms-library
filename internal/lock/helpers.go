package lock

import (
	"context"
	"time"

	"github.com/RacoonMediaServer/rms-library/v3/internal/model"
)

func TimedLock(ctx context.Context, lock Locker, id model.ID, timeout time.Duration) (Unlocker, error) {
	tCtx, tCancel := context.WithTimeout(ctx, timeout)
	defer tCancel()

	return lock.ContextLock(tCtx, id)
}
