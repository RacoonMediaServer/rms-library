package service

import (
	"context"
	"fmt"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/movies"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

func convertMovieInfo(in *models.SearchMoviesResult) *rms_library.MovieInfo {
	out := &rms_library.MovieInfo{
		Title:       *in.Title,
		Description: in.Description,
		Year:        uint32(in.Year),
		Poster:      in.Poster,
		Genres:      in.Genres,
		Rating:      float32(in.Rating),
	}

	if in.Seasons != 0 {
		seasons := uint32(in.Seasons)
		out.Seasons = &seasons
	}

	if in.Type == "tv-series" {
		out.Type = rms_library.MovieType_TvSeries
	} else {
		out.Type = rms_library.MovieType_Film
	}

	return out
}

func (l LibraryService) SearchMovie(ctx context.Context, request *rms_library.SearchMovieRequest, response *rms_library.SearchMovieResponse) error {
	logger.Infof("SearchMovie: %s", request.Text)

	limit := int64(request.Limit)
	q := &movies.SearchMoviesParams{
		Limit:   &limit,
		Q:       request.Text,
		Context: ctx,
	}

	resp, err := l.cli.Movies.SearchMovies(q, l.auth)
	if err != nil {
		err = fmt.Errorf("search movie failed: %w", err)
		logger.Error(err)
		return err
	}

	response.Movies = make([]*rms_library.FoundMovie, 0, len(resp.Payload.Results))
	for _, r := range resp.Payload.Results {
		mov := &rms_library.FoundMovie{
			Id:                *r.ID,
			Info:              convertMovieInfo(r),
			SeasonsDownloaded: make([]uint32, 0),
		}

		l.c.PutMovieInfo(mov.Id, mov.Info)

		seasons, err := l.db.GetDownloadedSeasons(ctx, *r.ID)
		if err != nil {
			logger.Warnf("load info about downloaded seasons failed: %s", err)
		}
		for i := range seasons {
			mov.SeasonsDownloaded = append(mov.SeasonsDownloaded, uint32(seasons[i]))
		}
		response.Movies = append(response.Movies, mov)
	}

	return nil
}
