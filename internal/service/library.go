package service

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
)

func convertMovie(mov *model.Movie) *rms_library.Movie {
	res := &rms_library.Movie{
		Id:   mov.ID,
		Info: &mov.Info,
	}
	if mov.Info.Type == rms_library.MovieType_Film {
		res.Film = &rms_library.FilmLayout{
			TorrentID: mov.TorrentID,
		}
		for _, f := range mov.Files {
			res.Film.Files = append(res.Film.Files, f.Path) // TODO: корректировка путей
		}
		return res
	}
	res.TvSeries = &rms_library.TvSeriesLayout{}
	for no, s := range mov.Seasons {
		l := rms_library.TvSeriesLayout_Season{}
		for _, e := range s.Episodes {
			l.Files = append(l.Files, e.Path) // TODO: корректировка путей
		}
		res.TvSeries.Seasons[uint32(no)] = &l
	}
	return res
}

func (l LibraryService) GetMovie(ctx context.Context, request *rms_library.GetMovieRequest, response *rms_library.GetMovieResponse) error {
	mov, err := l.db.GetMovie(ctx, request.ID)
	if err != nil {
		logger.Errorf("Cannot load movie from database: %s", err)
		return err
	}

	// если нет в библиотеке, то берем инфу из кеша
	if mov == nil {
		info, err := l.db.GetMovieInfo(ctx, request.ID)
		if err != nil {
			logger.Errorf("Cannot load movie from database: %s", err)
			return err
		}
		if info == nil {
			return nil
		}

		mov = &model.Movie{
			ID:   request.ID,
			Info: *info,
		}
	}

	response.Result = convertMovie(mov)
	return nil
}
