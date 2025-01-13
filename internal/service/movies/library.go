package movies

import (
	"context"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"go-micro.dev/v4/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (l LibraryService) Get(ctx context.Context, request *rms_library.GetMovieRequest, response *rms_library.GetMovieResponse) error {
	logger.Infof("GetMovie: %s", request.ID)
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

	response.Result = &rms_library.Movie{
		Id:   mov.ID,
		Info: &mov.Info,
		Size: l.dm.GetMovieStoreSize(ctx, mov),
	}

	return nil
}

func (l LibraryService) List(ctx context.Context, request *rms_library.GetMoviesRequest, response *rms_library.GetMoviesResponse) error {
	logger.Infof("GetMovies")
	movies, err := l.db.SearchMovies(ctx, request.Type)
	if err != nil {
		err = fmt.Errorf("load movies failed: %w", err)
		logger.Error(err)
		return err
	}

	response.Result = make([]*rms_library.Movie, 0, len(movies))
	for _, m := range movies {
		result := &rms_library.Movie{
			Id:   m.ID,
			Info: &m.Info,
			Size: l.dm.GetMovieStoreSize(ctx, m),
		}
		response.Result = append(response.Result, result)
	}
	return nil
}

func (l LibraryService) Delete(ctx context.Context, request *rms_library.DeleteRequest, empty *emptypb.Empty) error {
	logger.Infof("DeleteMovie: %s", request.ID)
	mov, err := l.db.GetMovie(ctx, request.ID)
	if err != nil {
		err = fmt.Errorf("load movie failed: %w", err)
		logger.Error(err)
		return err
	}
	if mov == nil {
		err = fmt.Errorf("movie %s not found", request.ID)
		logger.Warn(err)
		return err
	}

	// удаляем через менеджер закачек, чтобы удалить связанные загрузки
	if err = l.dm.RemoveMovie(ctx, mov); err != nil {
		err = fmt.Errorf("delete movie %s failed: %w", mov.Info.Title, err)
		logger.Error(err)
		return err
	}

	return nil
}
