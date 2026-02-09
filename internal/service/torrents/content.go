package torrents

import (
	"context"

	"github.com/RacoonMediaServer/rms-library/internal/model"
)

func (s *Service) download(ctx context.Context, item *model.ListItem, torrentLink *string, content []byte) error {
	var err error
	if len(content) == 0 {
		content, err = s.Movies.GetTorrentContent(ctx, *torrentLink)
		if err != nil {
			return err
		}
	}
	return s.Downloads.Download(ctx, item, content)
}
