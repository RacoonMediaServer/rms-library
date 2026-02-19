package movsearch

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/v3/pkg/selector"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type SeasonStrategy struct {
	Engine   SearchEngine
	Selector selector.MediaSelector
	SeasonNo uint
}

func (s SeasonStrategy) Search(ctx context.Context, id string, info *rms_library.MovieInfo, selopts selector.Options) ([]Result, error) {
	torrents, err := s.Engine.SearchTorrents(ctx, id, info, &s.SeasonNo)
	if err != nil {
		return nil, err
	}

	result := s.Selector.Select(torrents, selopts)
	torrentFile, err := s.Engine.GetTorrentFile(ctx, *result.Link)
	if err != nil {
		return nil, err
	}

	return []Result{{Torrent: torrentFile, Seasons: extractSeasonsFromResult(result)}}, nil
}
