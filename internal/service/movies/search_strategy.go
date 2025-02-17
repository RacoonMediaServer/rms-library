package movies

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/selector"
)

type foundTorrent struct {
	torrent []byte
	seasons []int64
}

func (t foundTorrent) Seasons() map[uint32]struct{} {
	result := map[uint32]struct{}{}
	for _, s := range t.seasons {
		result[uint32(s)] = struct{}{}
	}
	return result
}

func getSeasonsCount(torrents []foundTorrent) map[uint32]struct{} {
	result := map[uint32]struct{}{}
	for _, t := range torrents {
		for _, s := range t.seasons {
			result[uint32(s)] = struct{}{}
		}
	}
	return result
}

type searchStrategy interface {
	Search(ctx context.Context, mov *model.Movie, selopts selector.Options) ([]foundTorrent, error)
}

type simpleSearchStrategy struct {
	searchEngine torrentSearchEngine
	sel          selector.MediaSelector
}

func (s simpleSearchStrategy) Search(ctx context.Context, mov *model.Movie, selopts selector.Options) ([]foundTorrent, error) {
	torrents, err := s.searchEngine.SearchTorrents(ctx, mov, nil)
	if err != nil {
		return nil, err
	}

	result := s.sel.Select(torrents, selopts)
	torrentFile, err := s.searchEngine.GetTorrentFile(ctx, *result.Link)
	if err != nil {
		return nil, err
	}

	return []foundTorrent{foundTorrent{torrent: torrentFile}}, nil
}

type seasonSearchStrategy struct {
	searchEngine torrentSearchEngine
	sel          selector.MediaSelector
	seasonNo     uint32
}

func (s seasonSearchStrategy) Search(ctx context.Context, mov *model.Movie, selopts selector.Options) ([]foundTorrent, error) {
	torrents, err := s.searchEngine.SearchTorrents(ctx, mov, &s.seasonNo)
	if err != nil {
		return nil, err
	}

	result := s.sel.Select(torrents, selopts)
	torrentFile, err := s.searchEngine.GetTorrentFile(ctx, *result.Link)
	if err != nil {
		return nil, err
	}

	return []foundTorrent{foundTorrent{torrent: torrentFile, seasons: result.Seasons}}, nil
}

type fullSearchStrategy struct {
	searchEngine torrentSearchEngine
	sel          selector.MediaSelector
}

func (s fullSearchStrategy) Search(ctx context.Context, mov *model.Movie, selopts selector.Options) ([]foundTorrent, error) {
	result := []foundTorrent{}
	seasons := map[uint32]struct{}{}

	seasonSearcher := seasonSearchStrategy{searchEngine: s.searchEngine, sel: s.sel, seasonNo: 1}
	simpleSearcher := simpleSearchStrategy{searchEngine: s.searchEngine, sel: s.sel}

	skipAtOnce := false
	// если нужно быстро - выкачиваем сразу первый сезон
	if selopts.Criteria == selector.CriteriaFastest {
		season1, err := seasonSearcher.Search(ctx, mov, selopts)
		if err == nil {
			seasons = season1[0].Seasons()
			result = append(result, season1...)
			skipAtOnce = true
		}
	}

	// пробуем выкачать все
	if !skipAtOnce {
		compactOpts := selopts
		compactOpts.Criteria = selector.CriteriaCompact
		found, err := simpleSearcher.Search(ctx, mov, compactOpts)
		if err == nil {
			seasons = getSeasonsCount(found)
			result = append(result, found...)
		}
	}

	// пробуем выкачать все, что не удалось найти
	for season := uint32(1); season <= *mov.Info.Seasons; season++ {
		_, found := seasons[season]
		if found {
			continue
		}
		seasonSearcher.seasonNo = season
		torrents, err := seasonSearcher.Search(ctx, mov, selopts)
		if err == nil {
			result = append(result, torrents...)
			detectedSeasons := getSeasonsCount(torrents)
			for s := range detectedSeasons {
				seasons[s] = struct{}{}
			}
		}
	}

	if len(result) == 0 {
		return nil, errAnyTorrentsNotFound
	}

	return result, nil
}

type searchMissedStrategy struct {
	searchEngine torrentSearchEngine
	sel          selector.MediaSelector
	existing     map[uint]struct{}
}

func (s searchMissedStrategy) Search(ctx context.Context, mov *model.Movie, selopts selector.Options) ([]foundTorrent, error) {
	result := []foundTorrent{}

	seasonSearcher := seasonSearchStrategy{searchEngine: s.searchEngine, sel: s.sel, seasonNo: 1}

	// пробуем выкачать все, что не удалось найти
	for season := uint32(1); season <= *mov.Info.Seasons; season++ {
		_, found := s.existing[uint(season)]
		if found {
			continue
		}
		seasonSearcher.seasonNo = season
		torrents, err := seasonSearcher.Search(ctx, mov, selopts)
		if err == nil {
			result = append(result, torrents...)
			detectedSeasons := getSeasonsCount(torrents)
			for ds := range detectedSeasons {
				s.existing[uint(ds)] = struct{}{}
			}
		}
	}

	if len(result) == 0 {
		return nil, errAnyTorrentsNotFound
	}

	return result, nil
}
