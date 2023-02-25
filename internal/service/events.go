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

	mov, err := l.db.FindMovieByTorrentID(ctx, *event.TorrentID)
	if err != nil {
		logger.Warnf("Find movie by torrent ID failed: %s", err)
		return nil
	}

	if mov != nil {
		if event.Kind == events.Notification_DownloadComplete {
			logger.Infof("Movie '%s' downloaded, creating layout", mov.Info.Title)
			_ = l.m.CreateMovieLayout(mov)
			return nil
		}

		if event.Kind == events.Notification_TorrentRemoved {
			toDelete := false
			if mov.TorrentID == *event.TorrentID {
				logger.Infof("Torrent %s removed, so drop movie %s", *event.TorrentID, mov.ID)
				toDelete = true
			}

			no, ok := mov.FindSeasonByTorrentID(*event.TorrentID)
			if ok {
				logger.Infof("Torrent %s removed, so drop season %d of %s", *event.TorrentID, no, mov.ID)
				delete(mov.Seasons, no)
				toDelete = len(mov.Seasons) == 0
			}

			if toDelete {
				err = l.db.DeleteMovie(ctx, mov.ID)
				if err == nil {
					_ = l.m.DeleteMovieLayout(mov)
				}
			} else {
				err = l.db.UpdateMovieContent(mov)
				if err == nil {
					_ = l.m.CreateMovieLayout(mov)
				}
			}

			if err != nil {
				logger.Errorf("Movie %s proceed deletion failed: %s", mov.ID, err)
			}
		}
	}

	return nil
}
