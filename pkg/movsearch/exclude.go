package movsearch

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type ExcludeStrategy struct {
	Engine   SearchEngine
	Selector selector.MediaSelector
	Exclude  Seasons
}

func (s ExcludeStrategy) Search(ctx context.Context, id string, info *rms_library.MovieInfo, selopts selector.Options) ([]Result, error) {
	result := []Result{}

	seasonSearcher := SeasonStrategy{Engine: s.Engine, Selector: s.Selector, SeasonNo: 1}

	// пробуем выкачать все, что не удалось найти
	for season := uint(1); season <= uint(*info.Seasons); season++ {
		_, found := s.Exclude[season]
		if found {
			continue
		}
		seasonSearcher.SeasonNo = season
		torrents, err := seasonSearcher.Search(ctx, id, info, selopts)
		if err == nil {
			result = append(result, torrents...)
			detectedSeasons := GetMultipleResultsSeasons(torrents)
			for seasonNo := range detectedSeasons {
				s.Exclude[seasonNo] = struct{}{}
			}
		}
	}

	if len(result) == 0 {
		return nil, ErrAnyTorrentsNotFound
	}

	return result, nil
}
