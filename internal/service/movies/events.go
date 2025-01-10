package movies

import (
	"context"

	"github.com/RacoonMediaServer/rms-packages/pkg/events"
	"github.com/RacoonMediaServer/rms-packages/pkg/pubsub"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/server"
)

// Subscribe subscribes to notification events for monitoring downloads
func (l LibraryService) Subscribe(server server.Server) error {
	return micro.RegisterSubscriber(pubsub.NotificationTopic, server, l.handleNotification)
}

func (l LibraryService) handleNotification(ctx context.Context, event events.Notification) error {
	if event.TorrentID == nil {
		logger.Warn("Got notification without torrent ID")
		return nil
	}

	id, ok := l.dm.GetMovieByTorrent(*event.TorrentID)
	if !ok {
		logger.Warnf("Movie associated with torrent %s not found", *event.TorrentID)
		return nil
	}

	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		logger.Warnf("Find movie by torrent ID failed: %s", err)
		return nil
	}

	if mov == nil {
		logger.Warnf("Movie %s not found in the database", id)
		return nil
	}

	l.dm.HandleTorrentEvent(event.Kind, *event.TorrentID, mov)
	return nil
}
