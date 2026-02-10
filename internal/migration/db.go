package migration

import (
	"context"
	"fmt"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/db"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/logger"
)

type migratorFn func(f servicemgr.ServiceFactory) error

func (m *Migrator) migrateDatabaseV0ToV1(f servicemgr.ServiceFactory) error {
	torrCli := f.NewTorrent(false)

	movies, err := m.Database.SearchMovies(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("get all movies failed: %w", err)
	}

	for _, mov := range movies {
		if err = updateMovieLocations(m.Database, torrCli, mov); err != nil {
			return err
		}
	}

	return nil
}

func updateMovieLocations(d *db.Database, cli rms_torrent.RmsTorrentService, mov *model.Movie) error {
	updateDb := false
	for i := range mov.Torrents {
		t := &mov.Torrents[i]
		if t.Online || t.Location != "" {
			continue
		}
		info, err := cli.GetTorrentInfo(context.Background(), &rms_torrent.GetTorrentInfoRequest{Id: t.ID})
		if err != nil {
			logger.Warnf("Get info about torrent '%s' of '%s' failed: %s", t.ID, mov.Info.Title, err)
			continue
		}
		t.Location = info.Location
		updateDb = true
	}

	if updateDb {
		return d.UpdateMovieContent(context.Background(), mov)
	}

	return nil
}

func (m *Migrator) migrateDatabaseV1ToV2(f servicemgr.ServiceFactory) error {
	listFromTorrents := func(torrents []model.TorrentRecord) rms_library.List {
		for _, t := range torrents {
			if !t.Online {
				return rms_library.List_Favourites
			}
		}
		return rms_library.List_WatchList
	}

	dbOld, err := connectV1(m.Config.Database)
	if err != nil {
		return fmt.Errorf("connect to db failed: %s", err)
	}

	watchList, err := dbOld.getWatchList(context.Background())
	if err != nil {
		return fmt.Errorf("load watchlist failed: %s", err)
	}

	movies, err := dbOld.getMovies(context.Background())
	if err != nil {
		return fmt.Errorf("load movies failed: %s", err)
	}

	moviesMap := map[model.ID]*model.Movie{}
	for _, movV1 := range movies {
		mov := model.Movie{
			ListItem: model.ListItem{
				ID:          model.MakeID(movV1.ID, rms_library.ContentType_TypeMovies),
				CreatedAt:   time.Now(),
				Title:       movV1.Info.Title,
				List:        listFromTorrents(movV1.Torrents),
				Category:    model.GetVideoCategory(movV1.Info.Type),
				ContentType: rms_library.ContentType_TypeMovies,
				Torrents:    movV1.Torrents,
			},
			Info:  movV1.Info,
			Voice: movV1.Voice,
		}
		if err := m.Database.AddMovie(context.Background(), &mov); err != nil {
			logger.Warnf("Add movie '%s' [ %s ] failed: %s", mov.Title, mov.ID, err)
		} else {
			moviesMap[mov.ID] = &mov
		}
	}

	for _, wlItem := range watchList {
		id := model.MakeID(wlItem.ID, rms_library.ContentType_TypeMovies)
		mov, ok := moviesMap[id]
		if !ok {
			mov = &model.Movie{
				ListItem: model.ListItem{
					ID:          id,
					CreatedAt:   time.Now(),
					Title:       wlItem.MovieInfo.Title,
					List:        rms_library.List_Archive,
					Category:    model.GetVideoCategory(wlItem.MovieInfo.Type),
					ContentType: rms_library.ContentType_TypeMovies,
				},
				Info:             wlItem.MovieInfo,
				ArchivedTorrents: convertArchivedTorrents(wlItem.Torrents),
				ArchivedSeasons:  convertArchivedSeasons(wlItem.Seasons),
			}
			if err := m.Database.AddMovie(context.Background(), mov); err != nil {
				logger.Warnf("Add movie '%s' [ %s ] failed: %s", mov.Title, mov.ID, err)
			}
		} else {
			mov.ArchivedTorrents = convertArchivedTorrents(wlItem.Torrents)
			mov.ArchivedSeasons = convertArchivedSeasons(wlItem.Seasons)
			if err := m.Database.UpdateMovieArchiveContent(context.Background(), mov); err != nil {
				logger.Warnf("Update movie archive of '%s' [ %s ] failed: %s", mov.Title, mov.ID, err)
			}
		}
	}

	return nil
}

func convertArchivedSeasons(torrentItems map[uint][]torrentItemV1) map[uint][]model.TorrentSearchResult {
	result := map[uint][]model.TorrentSearchResult{}
	for season, torrents := range torrentItems {
		result[season] = convertArchivedTorrents(torrents)
	}
	return result
}

func convertArchivedTorrents(torrentItems []torrentItemV1) []model.TorrentSearchResult {
	result := make([]model.TorrentSearchResult, 0, len(torrentItems))
	for _, item := range torrentItems {
		conv := model.TorrentSearchResult{
			SearchTorrentsResult: item.SearchTorrentsResult,
			Path:                 item.TorrentContent,
		}
		result = append(result, conv)
	}
	return result
}
