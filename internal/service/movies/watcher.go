package movies

import (
	"context"
	"fmt"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/lock"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

const lockWait = 5 * time.Second

func isContentMissing(mov *model.Movie) bool {
	switch mov.List {
	case rms_library.List_Archive:
		if mov.Info.Type == rms_library.MovieType_Film {
			return len(mov.ArchivedTorrents) == 0
		} else if mov.Info.Type == rms_library.MovieType_TvSeries {
			return len(mov.ArchivedSeasons) == 0
		} else {
			return false
		}
	case rms_library.List_Favourites:
		for _, t := range mov.Torrents {
			if !t.Online {
				return false
			}
		}
		return true
	case rms_library.List_WatchList:
		for _, t := range mov.Torrents {
			if t.Online {
				return false
			}
		}
		return true
	}
	return false
}

func (l MoviesService) asyncWatch(log logger.Logger, ctx context.Context, id model.ID) error {
	lk, err := lock.TimedLock(ctx, l.lk, id, lockWait)
	if err != nil {
		return fmt.Errorf("Lock item failed: %w", err)
	}
	defer lk.Unlock()

	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return fmt.Errorf("load movie from database failed: %w", err)
	}

	// проверяем различные проблемы и рассинхрон
	// каждую найденную ситуацию разруливаем

	// 1) если какие то торренты пропали с диска - корректируем внутреннее хранилище
	if mov.List != rms_library.List_Archive {
		l.watcherRemoveMissedTorrents(log, ctx, mov)
	}

	// 2) сихронизируем контент и тип списка, в который добавлен item - удаляем торренты, которые считаем лишними
	l.watcherRemoveUnusedTorrents(log, ctx, mov)

	// 3) запускаем загрузку если полностью отсутствует контент
	if isContentMissing(mov) {
		log.Logf(logger.WarnLevel, "Content is missing, try to download all")
		return l.downloadContent(log, ctx, mov)
	}

	// 4) синхронизируем информацию о торрентах
	l.watcherSyncTorrentInfo(log, ctx, mov)

	return nil
}

func (l MoviesService) watcherRemoveUnusedTorrents(log logger.Logger, ctx context.Context, mov *model.Movie) {
	removeUnusedTorrent := func(log logger.Logger, ctx context.Context, mov *model.Movie, t *model.TorrentRecord) {
		log.Logf(logger.DebugLevel, "Unused torrent found: %s [ %s ]", t.Title, t.ID)
		if err := l.dm.RemoveTorrent(ctx, &mov.ListItem, t.ID); err != nil {
			log.Logf(logger.WarnLevel, "Remove unused torrent failed: %s", err)
		}
	}
	switch mov.List {
	case rms_library.List_Archive:
		for _, t := range mov.Torrents {
			removeUnusedTorrent(log, ctx, mov, &t)
		}
	case rms_library.List_WatchList:
		for _, t := range mov.Torrents {
			if !t.Online {
				removeUnusedTorrent(log, ctx, mov, &t)
			}
		}
	case rms_library.List_Favourites:
		for _, t := range mov.Torrents {
			if t.Online {
				removeUnusedTorrent(log, ctx, mov, &t)
			}
		}
	}
}

func (l MoviesService) watcherRemoveMissedTorrents(log logger.Logger, ctx context.Context, mov *model.Movie) {
	if err := l.dm.DropMissedTorrents(ctx, &mov.ListItem); err != nil {
		log.Logf(logger.WarnLevel, "Drop missed torrents failed: %s", err)
	}
}

func (l MoviesService) watcherSyncTorrentInfo(log logger.Logger, ctx context.Context, mov *model.Movie) {
	if err := l.dm.UpdateTorrentInfo(ctx, &mov.ListItem); err != nil {
		log.Logf(logger.WarnLevel, "Sync torrents infromation failed: %s", err)
	}
}
