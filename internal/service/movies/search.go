package movies

import (
	"context"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client/movies"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

func convertMovieInfo(in *models.SearchMoviesResult) *rms_library.MovieInfo {
	out := &rms_library.MovieInfo{
		Title:         *in.Title,
		Description:   in.Description,
		Year:          uint32(in.Year),
		Poster:        in.Poster,
		Genres:        in.Genres,
		Rating:        float32(in.Rating),
		OriginalTitle: in.OriginalTitle,
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

func (l MoviesService) Search(ctx context.Context, request *rms_library.MoviesSearchRequest, response *rms_library.MoviesSearchResponse) error {
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
	logger.Infof("Got %d results", len(resp.Payload.Results))

	response.Movies = make([]*rms_library.FoundMovie, 0, len(resp.Payload.Results))
	for _, r := range resp.Payload.Results {
		id := model.MakeID(*r.ID, rms_library.ContentType_TypeMovies)
		mov := &rms_library.FoundMovie{
			Id:   id.String(),
			Info: convertMovieInfo(r),
		}

		if err = l.db.PutMovieInfo(ctx, id, mov.Info); err != nil {
			logger.Warnf("Save movie info to cache failed: %s", err)
		}

		response.Movies = append(response.Movies, mov)
	}

	return nil
}
