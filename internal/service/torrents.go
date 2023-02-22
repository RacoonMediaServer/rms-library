package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/analysis"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"go-micro.dev/v4/logger"
	"sort"
)

var errAnyTorrentsNotFound = errors.New("any torrents not found")

func (l LibraryService) searchTorrents(ctx context.Context, mov *rms_library.MovieInfo, season *uint32) ([]*models.SearchTorrentsResult, error) {
	limit := int64(searchTorrentsLimit)
	torrentType := "movies"
	strong := true
	q := &torrents.SearchTorrentsParams{
		Limit:   &limit,
		Q:       mov.Title,
		Season:  nil,
		Type:    &torrentType,
		Year:    nil,
		Context: ctx,
		Strong:  &strong,
	}

	if mov.Type == rms_library.MovieType_TvSeries && season != nil {
		s := int64(*season)
		q.Season = &s
	}

	if mov.Type == rms_library.MovieType_Film && mov.Year != 0 {
		year := int64(mov.Year)
		q.Year = &year
	}

	resp, err := l.cli.Torrents.SearchTorrents(q, l.auth)
	if err != nil {
		return nil, err
	}
	return resp.Payload.Results, nil
}

func sortTorrentMovies(list []*models.SearchTorrentsResult) {
	qualityPrior := map[string]int{
		"1080p": 4,
		"720p":  3,
		"480p":  2,
		"":      1,
	}
	sort.SliceStable(list, func(i, j int) bool {
		return qualityPrior[list[i].Quality] > qualityPrior[list[j].Quality]
	})
}

func (l LibraryService) searchAndDownloadMovie(ctx context.Context, mov *model.Movie, season *uint32) error {
	list, err := l.searchTorrents(ctx, &mov.Info, season)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errAnyTorrentsNotFound
	}

	sortTorrentMovies(list)

	resp, err := l.downloadTorrent(ctx, list[0])
	if err != nil {
		return err
	}

	for _, file := range resp.Files {
		result := analysis.Analyze(file)
		f := model.File{
			Path:  file,
			Title: result.EpisodeName,
			Type:  result.FileType,
			No:    result.Episode,
		}
		mov.AddFile(resp.Id, f, result.Season)
	}

	if err = l.db.UpdateMovieContent(mov); err != nil {
		l.removeTorrent(resp.Id)
		return err
	}

	if err = l.m.CreateMovieLayout(mov); err != nil {
		logger.Warnf("Create storage layout failed: %s", err)
	}

	return nil
}

func (l LibraryService) DownloadMovie(ctx context.Context, request *rms_library.DownloadMovieRequest, response *rms_library.DownloadMovieResponse) error {
	var downloadedSeasons []uint32

	logger.Infof("DownloadMovie: %s", request.Id)
	// 1. Вытаскиваем из кеша инфу о медиа
	movInfo, err := l.db.GetMovieInfo(ctx, request.Id)
	if err != nil {
		logger.Errorf("Get movie info from cache failed: %s", err)
		return err
	}

	if movInfo == nil {
		err := fmt.Errorf("movie '%s' not found in the cache", request.Id)
		logger.Warn(err)
		return err
	}

	// 2. Создаем или вытаскиваем существующую инфу из базы о медиа
	mov := &model.Movie{
		ID:   request.Id,
		Info: *movInfo,
	}
	if err := l.db.GetOrCreateMovie(ctx, mov); err != nil {
		err := fmt.Errorf("database error: %w", err)
		logger.Error(err)
		return err
	}

	// 3. Создаем список сезонов для скачивания
	var seasons []uint32
	if movInfo.Type == rms_library.MovieType_TvSeries {
		if request.Season == nil {
			if movInfo.Seasons != nil {
				for i := 1; i <= int(*movInfo.Seasons); i++ {
					if !mov.IsSeasonDownloaded(uint(i)) {
						seasons = append(seasons, uint32(i))
					}
				}
			}
		} else {
			seasons = append(seasons, *request.Season)
		}

		if len(seasons) == 0 {
			logger.Warnf("Cannot find any season for '%s'", movInfo.Title)
			return nil
		}

		// скачиваем все сезоны
		for _, s := range seasons {
			if err := l.searchAndDownloadMovie(ctx, mov, &s); err != nil {
				logger.Errorf("Cannot download season #%d of '%s': %s", s, movInfo.Title, err)
				continue
			}
			downloadedSeasons = append(downloadedSeasons, s)
		}

		if len(downloadedSeasons) == 0 {
			return errors.New("cannot download anything")
		}

		response.Found = true
		response.Seasons = downloadedSeasons
	} else {
		if err := l.searchAndDownloadMovie(ctx, mov, nil); err != nil {
			if errors.Is(err, errAnyTorrentsNotFound) {
				return nil
			}
			logger.Error(err)
			return err
		}
		response.Found = true
	}

	return nil
}

func (l LibraryService) downloadTorrent(ctx context.Context, t *models.SearchTorrentsResult) (*rms_torrent.DownloadResponse, error) {
	download := &torrents.DownloadTorrentParams{
		Link:    *t.Link,
		Context: ctx,
	}
	buf := bytes.NewBuffer([]byte{})

	_, err := l.cli.Torrents.DownloadTorrent(download, l.auth, buf)
	if err != nil {
		return nil, fmt.Errorf("download torrent file failed: %w", err)
	}

	service := l.f.NewTorrent()
	resp, err := service.Download(ctx, &rms_torrent.DownloadRequest{What: buf.Bytes()})
	if err != nil {
		return nil, fmt.Errorf("push torrent file to queue failed: %w", err)
	}
	return resp, nil
}

func (l LibraryService) removeTorrent(id string) {
	service := l.f.NewTorrent()
	_, err := service.RemoveTorrent(context.Background(), &rms_torrent.RemoveTorrentRequest{Id: id})
	if err != nil {
		logger.Errorf("Remove torrent failed: %s", err)
	}
}
