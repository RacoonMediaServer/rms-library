package movies

import (
	"context"
	"fmt"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/movies"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/torrents"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

const checkAvailableInterval = 24 * time.Hour
const updatesRefreshInterval = 24 * time.Hour * 7

func (l LibraryService) checkAvailableUpdates() {
	for {
		<-time.After(checkAvailableInterval)

		typeSeries := rms_library.MovieType_TvSeries
		series, err := l.db.SearchMovies(context.Background(), &typeSeries)
		if err != nil {
			err = fmt.Errorf("search movies in database failed: %s", err)
			logger.Error(err)
		}

		for _, mov := range series {
			if time.Since(mov.LastAvailableCheck) >= updatesRefreshInterval {
				if err = l.findUpdates(mov); err != nil {
					logger.Warnf("Find movie '%s' updates failed: %s", mov.Info.Title, err)
				}
			}
		}
	}
}

func (l LibraryService) findUpdates(mov *model.Movie) error {
	info, err := l.cli.Movies.GetMovieInfo(&movies.GetMovieInfoParams{ID: mov.ID, Context: context.Background()}, l.auth)
	if err != nil {
		return err
	}
	if info.Payload.Seasons == 0 {
		return nil
	}

	mov.LastAvailableCheck = time.Now()
	mov.AvailableSeasons = nil

	limit := int64(1)
	strong := true
	torrentType := "movies"

	for no := uint(1); no <= uint(info.Payload.Seasons); no++ {
		if _, ok := mov.Seasons[no]; ok {
			continue
		}
		// проверяем что сезон есть на торрентах
		season := int64(no)
		result, err := l.cli.Torrents.SearchTorrents(&torrents.SearchTorrentsParams{
			Limit:   &limit,
			Q:       mov.Info.Title,
			Season:  &season,
			Strong:  &strong,
			Type:    &torrentType,
			Context: context.Background(),
		}, l.auth)
		if err != nil {
			logger.Warnf("Find available torrent failed: %s", err)
			continue
		}
		if len(result.Payload.Results) == 0 {
			continue
		}
		logger.Infof("Found season %d for '%s'", no, mov.Info.Title)

		mov.AvailableSeasons = append(mov.AvailableSeasons, no)
	}

	if err = l.db.UpdateAvailableContent(context.Background(), mov); err != nil {
		return fmt.Errorf("update available content in database failed: %w", err)
	}

	return nil
}

func (l LibraryService) GetTvSeriesUpdates(ctx context.Context, empty *emptypb.Empty, response *rms_library.GetTvSeriesUpdatesResponse) error {
	typeSeries := rms_library.MovieType_TvSeries
	series, err := l.db.SearchMovies(context.Background(), &typeSeries)
	if err != nil {
		err = fmt.Errorf("search movies in database failed: %s", err)
		logger.Error(err)
		return err
	}
	for _, mov := range series {
		seasons := make([]uint32, 0, len(mov.AvailableSeasons))
		for _, no := range mov.AvailableSeasons {
			if _, ok := mov.Seasons[no]; !ok {
				seasons = append(seasons, uint32(no))
			}
		}

		if len(seasons) == 0 {
			continue
		}

		response.Updates = append(response.Updates, &rms_library.TvSeriesUpdate{
			Id:               mov.ID,
			Info:             &mov.Info,
			SeasonsAvailable: seasons,
		})
	}

	return nil
}
