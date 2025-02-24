package movies

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/movsearch"
	"github.com/RacoonMediaServer/rms-library/pkg/selector"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/media"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

var errAnyTorrentsNotFound = errors.New("any torrents not found")

func (l LibraryService) getOrCreateMovie(ctx context.Context, id string, canUseWatchList bool) (*model.Movie, error) {
	// 1. Вытаскиваем из кеша инфу о медиа
	movInfo, err := l.db.GetMovieInfo(ctx, id)
	if err != nil {
		return nil, err
	}

	if movInfo == nil {
		if canUseWatchList {
			item, err := l.db.GetWatchListItem(ctx, id)
			if err != nil {
				return nil, err
			}
			if item != nil {
				movInfo = &item.MovieInfo
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
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

func (l LibraryService) DownloadAuto(ctx context.Context, request *rms_library.DownloadMovieAutoRequest, response *rms_library.DownloadMovieAutoResponse) error {
	logger.Infof("DownloadMovieAuto: %s", request.Id)
	mov, err := l.getOrCreateMovie(ctx, request.Id, request.UseWatchList)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}
	defer l.removeMovieIfEmpty(ctx, request.Id)

	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)
	if request.UseWatchList {
		searchByWatchList := newWatchListSearchEngine(l.db, l.dir)
		searchByWatchList.SetNext(searchEngine)
		searchEngine = searchByWatchList
	}

	var strategy movsearch.Strategy
	sel := l.getMovieSelector(mov)

	if mov.Info.Type == rms_library.MovieType_TvSeries {
		if request.Season == nil {
			existsSeasons := l.dir.GetDownloadedSeasons(mov)
			if len(existsSeasons) == 0 {
				strategy = &movsearch.FullStrategy{Engine: searchEngine, Selector: sel}
			} else {
				strategy = &movsearch.ExcludeStrategy{Engine: searchEngine, Selector: sel, Exclude: existsSeasons}
			}

		} else {
			strategy = &movsearch.SeasonStrategy{Engine: searchEngine, Selector: sel, SeasonNo: uint(*request.Season)}
		}
	} else {
		strategy = &movsearch.SimpleStrategy{Engine: searchEngine, Selector: sel}
	}

	selopts := selector.Options{
		Criteria:  selector.CriteriaQuality,
		MediaType: media.Movies,
		Query:     mov.Info.Title,
	}

	if request.Faster {
		selopts.Criteria = selector.CriteriaFastest
	}

	result, err := strategy.Search(ctx, mov.ID, &mov.Info, selopts)
	if err != nil {
		if errors.Is(err, errAnyTorrentsNotFound) {
			return nil
		}
		logger.Errorf("Search torrents failed: %s", err)
		return err
	}

	for _, r := range result {
		if err = l.dm.DownloadMovie(ctx, mov, "", r.Torrent, request.Faster); err != nil {
			logger.Errorf("add movie to download manager failed: %s", err)
		}
	}

	seasons := movsearch.GetMultipleResultsSeasons(result)
	for s := range seasons {
		response.Seasons = append(response.Seasons, uint32(s))
	}
	sort.SliceStable(response.Seasons, func(i, j int) bool { return response.Seasons[i] < response.Seasons[j] })
	response.Found = true

	return nil
}

func (l LibraryService) FindTorrents(ctx context.Context, request *rms_library.FindMovieTorrentsRequest, response *rms_library.FindTorrentsResponse) error {
	logger.Infof("FindMovieTorrents: %s", request.Id)

	mov, err := l.getOrCreateMovie(ctx, request.Id, request.UseWatchList)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}
	defer l.removeMovieIfEmpty(ctx, request.Id)

	searchEngine := movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth)
	if request.UseWatchList {
		searchByWatchList := newWatchListSearchEngine(l.db, l.dir)
		searchByWatchList.SetNext(searchEngine)
		searchEngine = searchByWatchList
	}

	var season *uint
	if request.Season != nil {
		season = new(uint)
		*season = uint(*request.Season)
	}
	resp, err := searchEngine.SearchTorrents(ctx, mov.ID, &mov.Info, season)
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
		l.torrentToResult[*t.Link] = t
	}
	return nil
}

func (l LibraryService) Download(ctx context.Context, request *rms_library.DownloadTorrentRequest, empty *emptypb.Empty) error {
	logger.Infof("DownloadTorrent: %s", request.TorrentId)
	mediaID, ok := l.torrentToMovieID[request.TorrentId]
	if !ok {
		err := errors.New("torrent link not found in the cache")
		logger.Warn(err)
		return err
	}
	torrent := l.torrentToResult[request.TorrentId]

	mov, err := l.getOrCreateMovie(ctx, mediaID, false)
	if err != nil {
		err = fmt.Errorf("get or create movie failed: %s", err)
		logger.Error(err)
		return err
	}
	defer l.removeMovieIfEmpty(ctx, mediaID)

	searchEngine := newWatchListSearchEngine(l.db, l.dir)
	searchEngine.SetNext(movsearch.NewRemoteSearchEngine(l.cli.Torrents, l.auth))

	data, err := searchEngine.GetTorrentFile(ctx, *torrent.Link)
	if err != nil {
		logger.Errorf("Download torrent failed: %s", err)
		return err
	}

	return l.dm.DownloadMovie(ctx, mov, torrent.Voice, data, false)
}

func (l LibraryService) removeMovieIfEmpty(ctx context.Context, id string) {
	mov, err := l.db.GetMovie(ctx, id)
	if err != nil {
		return
	}
	if len(mov.Torrents) == 0 {
		logger.Debugf("Removing empty movie record: %s [ %s ]", id, mov.Info.Title)
		if err = l.db.DeleteMovie(ctx, id); err != nil {
			logger.Warnf("Remove empty movie '%s' failed: %s", id, err)
		}
	}
}

func (l LibraryService) Upload(ctx context.Context, request *rms_library.UploadMovieRequest, empty *emptypb.Empty) error {
	mov := model.Movie{
		ID:   request.Id,
		Info: *request.Info,
	}

	err := l.db.GetOrCreateMovie(ctx, &mov)
	if err != nil {
		logger.Errorf("Store movie info failed: %s", err)
		return err
	}
	defer l.removeMovieIfEmpty(ctx, request.Id)

	if err = l.dm.DownloadMovie(ctx, &mov, "", request.TorrentFile, false); err != nil {
		logger.Errorf("Start download given file failed: %s", err)
		return err
	}

	return nil
}
