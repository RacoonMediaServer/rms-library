package movsearch

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type FullStrategy struct {
	Engine   SearchEngine
	Selector selector.MediaSelector
}

func (s FullStrategy) Search(ctx context.Context, id string, info *rms_library.MovieInfo, selopts selector.Options) ([]Result, error) {
	result := []Result{}
	seasons := Seasons{}

	seasonSearcher := SeasonStrategy{Engine: s.Engine, Selector: s.Selector, SeasonNo: 1}
	simpleSearcher := SimpleStrategy{Engine: s.Engine, Selector: s.Selector}

	skipAtOnce := false
	// если нужно быстро - выкачиваем сразу первый сезон
	if selopts.Criteria == selector.CriteriaFastest {
		season1, err := seasonSearcher.Search(ctx, id, info, selopts)
		if err == nil {
			seasons = season1[0].Seasons
			result = append(result, season1...)
			skipAtOnce = true
		}
	}

	// пробуем выкачать все
	if !skipAtOnce {
		compactOpts := selopts
		compactOpts.Criteria = selector.CriteriaCompact
		found, err := simpleSearcher.Search(ctx, id, info, compactOpts)
		if err == nil {
			seasons = GetMultipleResultsSeasons(found)
			result = append(result, found...)
		}
	}

	// пробуем выкачать все, что не удалось найти
	if info.Seasons != nil {
		for season := uint(1); season <= uint(*info.Seasons); season++ {
			_, found := seasons[season]
			if found {
				continue
			}
			seasonSearcher.SeasonNo = season
			torrents, err := seasonSearcher.Search(ctx, id, info, selopts)
			if err == nil {
				result = append(result, torrents...)
				detectedSeasons := GetMultipleResultsSeasons(torrents)
				seasons.Union(detectedSeasons)
			}
		}
	}

	if len(result) == 0 {
		return nil, ErrAnyTorrentsNotFound
	}

	return result, nil
}
