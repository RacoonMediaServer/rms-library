package storage

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"path"
)

const (
	nameFilms    = "Фильмы"
	nameTvSeries = "Сериалы"
	nameByGenre  = "Жанры"
	nameByAlpha  = "Алфавит"
	nameByYear   = "Год"
)

func composeMovieFileName(mov *model.Movie, f *model.File) string {
	_, fileName := path.Split(f.Path)
	ext := path.Ext(f.Path)

	if mov.Info.Type == rms_library.MovieType_Film {
		if len(mov.Files) == 1 {
			return fmt.Sprintf("%s%s", mov.Info.Title, ext)
		}
		if f.Title == "" {
			return fileName
		}
		return escape(f.Title) + ext
	}

	if f.No < 0 {
		if f.Title == "" {
			return fileName
		}
		return f.Title + ext
	}
	if f.Title == "" {
		return fmt.Sprintf("E%02d%s", f.No, ext)
	}
	return fmt.Sprintf("E%02d. %s", f.No, fileName)
}

func getMovieCategoryDir(mov *model.Movie) string {
	switch mov.Info.Type {
	case rms_library.MovieType_TvSeries:
		return nameTvSeries
	case rms_library.MovieType_Film:
		return nameFilms
	default:
		return ""
	}
}
