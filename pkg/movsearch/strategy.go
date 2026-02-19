package movsearch

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/v3/pkg/selector"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type Strategy interface {
	Search(ctx context.Context, id string, info *rms_library.MovieInfo, selopts selector.Options) ([]Result, error)
}
