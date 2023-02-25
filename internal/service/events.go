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
	if event.Kind != events.Notification_DownloadComplete || event.TorrentID == nil {
		return nil
	}

	mov, err := l.db.FindMovieByTorrentID(ctx, *event.TorrentID)
	if err != nil {
		logger.Warnf("Find movie by torrent ID failed: %s", err)
		return nil
	}
	if mov != nil {
		logger.Infof("Movie '%s' downloaded, creating layout", mov.Info.Title)
		_ = l.m.CreateMovieLayout(mov)
	}

	return nil
}
