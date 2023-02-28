package service

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
		return nil
	}

	id, ok := l.dm.GetMovieByTorrent(*event.TorrentID)
	if !ok {
		return nil
	}

	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		logger.Warnf("Find movie by torrent ID failed: %s", err)
		return nil
	}

	if mov != nil {
		if event.Kind == events.Notification_DownloadComplete {
			logger.Infof("Movie '%s' downloaded, creating layout", mov.Info.Title)
			if err = l.dir.CreateMovieLayout(mov); err != nil {
				logger.Warnf("Create movie layout for %s failed: %s", mov.Info.Title, err)
			}
			return nil
		}

		if event.Kind == events.Notification_TorrentRemoved {
			if l.dm.RemoveMovieTorrent(*event.TorrentID, mov) {
				if err = l.db.DeleteMovie(context.Background(), mov.ID); err != nil {
					logger.Errorf("Delete movie %s from database failed: %s", mov.Info.Title, err)
				}
				return nil
			}
			if err = l.db.UpdateMovieContent(mov); err != nil {
				logger.Errorf("Update movie %s in database failed: %s", mov.Info.Title, err)
			}
		}
	}

	return nil
}
