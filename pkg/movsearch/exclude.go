package movsearch

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/v3/pkg/selector"
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

	seasonsKnown := info.Seasons != nil
	shouldContinueIterate := func(seasonNo uint) bool {
		if seasonsKnown {
			return seasonNo <= uint(*info.Seasons)
		}
		return false
	}

	for season := uint(1); shouldContinueIterate(season); season++ {
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
		} else if !seasonsKnown {
			break
		}
	}

	if len(result) == 0 {
		return nil, ErrAnyTorrentsNotFound
	}

	return removeDuplicatedResults(result), nil
}
