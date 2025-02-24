package movsearch

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type SimpleStrategy struct {
	Engine   SearchEngine
	Selector selector.MediaSelector
}

func (s SimpleStrategy) Search(ctx context.Context, id string, info *rms_library.MovieInfo, selopts selector.Options) ([]Result, error) {
	torrents, err := s.Engine.SearchTorrents(ctx, id, info, nil)
	if err != nil {
		return nil, err
	}

	result := s.Selector.Select(torrents, selopts)
	torrentFile, err := s.Engine.GetTorrentFile(ctx, *result.Link)
	if err != nil {
		return nil, err
	}

	return []Result{{Torrent: torrentFile}}, nil
}
