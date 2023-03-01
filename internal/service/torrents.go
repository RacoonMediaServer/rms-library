package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
	"sort"
)

var errAnyTorrentsNotFound = errors.New("any torrents not found")

func (l LibraryService) searchMovieTorrents(ctx context.Context, mov *rms_library.MovieInfo, season *uint32, limit uint) ([]*models.SearchTorrentsResult, error) {
	limitInt := int64(limit)
	torrentType := "movies"
	strong := true
	q := &torrents.SearchTorrentsParams{
		Limit:   &limitInt,
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
	// хотим в приоритете иметь 1080p, в дальнейшем следует вынести в настройки
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

func (l LibraryService) getOrCreateMovie(ctx context.Context, id string) (*model.Movie, error) {
	// 1. Вытаскиваем из кеша инфу о медиа
	movInfo, err := l.db.GetMovieInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	if movInfo == nil {
		return nil, err
	}

	// 2. Создаем или вытаскиваем существующую инфу из базы о медиа
	mov := &model.Movie{
		ID:   id,
		Info: *movInfo,
	}
	if err := l.db.GetOrCreateMovie(ctx, mov); err != nil {
		return nil, err
	}
	return mov, nil
}

func (l LibraryService) searchAndDownloadMovie(ctx context.Context, mov *model.Movie, season *uint32) error {
	list, err := l.searchMovieTorrents(ctx, &mov.Info, season, searchTorrentsLimit)
	if err != nil {
		return err
	}

	if len(list) == 0 {
		return errAnyTorrentsNotFound
	}

	sortTorrentMovies(list)

	return l.downloadMovie(ctx, mov, *list[0].Link)
}

func (l LibraryService) downloadMovie(ctx context.Context, mov *model.Movie, link string) error {
	torrent, err := l.downloadTorrent(ctx, link)
	if err != nil {
		return fmt.Errorf("download torrent file failed: %w", err)
	}

	// добавляем в менеджер загрузок
	if err = l.dm.DownloadMovie(ctx, mov, torrent); err != nil {
		return fmt.Errorf("add movie to download manager failed: %w", err)
	}

	return nil
}

func (l LibraryService) DownloadMovieAuto(ctx context.Context, request *rms_library.DownloadMovieAutoRequest, response *rms_library.DownloadMovieAutoResponse) error {
	var downloadedSeasons []uint32

	logger.Infof("DownloadMovieAuto: %s", request.Id)
	mov, err := l.getOrCreateMovie(ctx, request.Id)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}

	// создаем список сезонов для скачивания
	var seasons []uint32
	if mov.Info.Type == rms_library.MovieType_TvSeries {
		if request.Season == nil {
			if mov.Info.Seasons != nil {
				for i := 1; i <= int(*mov.Info.Seasons); i++ {
					if !mov.IsSeasonDownloaded(uint(i)) {
						seasons = append(seasons, uint32(i))
					}
				}
			}
		} else {
			seasons = append(seasons, *request.Season)
		}

		if len(seasons) == 0 {
			logger.Warnf("Cannot find any season for '%s'", mov.Info.Title)
			return nil
		}

		// скачиваем все сезоны
		for _, s := range seasons {
			if err := l.searchAndDownloadMovie(ctx, mov, &s); err != nil {
				logger.Errorf("Cannot download season #%d of '%s': %s", s, mov.Info.Title, err)
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

func (l LibraryService) downloadTorrent(ctx context.Context, link string) ([]byte, error) {
	download := &torrents.DownloadTorrentParams{
		Link:    link,
		Context: ctx,
	}
	buf := bytes.NewBuffer([]byte{})

	_, err := l.cli.Torrents.DownloadTorrent(download, l.auth, buf)
	if err != nil {
		return nil, fmt.Errorf("download torrent file failed: %w", err)
	}

	return buf.Bytes(), nil
}

func (l LibraryService) FindMovieTorrents(ctx context.Context, request *rms_library.FindMovieTorrentsRequest, response *rms_library.FindTorrentsResponse) error {
	logger.Infof("FindMovieTorrents: %s", request.Id)

	mov, err := l.getOrCreateMovie(ctx, request.Id)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}

	resp, err := l.searchMovieTorrents(ctx, &mov.Info, request.Season, uint(request.Limit))
	if err != nil {
		err = fmt.Errorf("search torrents failed: %s", err)
		logger.Error(err)
		return err
	}

	for _, t := range resp {
		response.Results = append(response.Results, &rms_library.Torrent{
			Id:      *t.Link,
			Title:   *t.Title,
			Size:    uint64(*t.Size),
			Seeders: uint32(*t.Seeders),
		})
		l.torrentToMovieID[*t.Link] = mov.ID
	}
	return nil
}

func (l LibraryService) DownloadTorrent(ctx context.Context, request *rms_library.DownloadTorrentRequest, empty *emptypb.Empty) error {
	logger.Infof("DownloadTorrent: %s", request.TorrentId)
	mediaID, ok := l.torrentToMovieID[request.TorrentId]
	if !ok {
		err := errors.New("torrent link not found in the cache")
		logger.Warn(err)
		return err
	}

	mov, err := l.getOrCreateMovie(ctx, mediaID)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}

	return l.downloadMovie(ctx, mov, request.TorrentId)
}

func (l LibraryService) FindTorrents(ctx context.Context, request *rms_library.FindTorrentsRequest, response *rms_library.FindTorrentsResponse) error {
	logger.Infof("FindTorrents: %s", request.Query)
	limitInt := int64(request.Limit)
	q := &torrents.SearchTorrentsParams{
		Limit:   &limitInt,
		Q:       request.Query,
		Context: ctx,
		Strong:  &request.Strong,
	}

	resp, err := l.cli.Torrents.SearchTorrents(q, l.auth)
	if err != nil {
		err = fmt.Errorf("search torrents failed: %w", err)
		logger.Error(err)
		return err
	}

	for _, t := range resp.Payload.Results {
		response.Results = append(response.Results, &rms_library.Torrent{
			Id:      *t.Link,
			Title:   *t.Title,
			Size:    uint64(*t.Size),
			Seeders: uint32(*t.Seeders),
		})
	}
	return nil
}
