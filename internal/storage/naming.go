package storage

import (
	"fmt"
	"path"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

const (
	nameFilms    = "Фильмы"
	nameTvSeries = "Сериалы"
	nameClips    = "Ролики"
	nameByGenre  = "Жанры"
	nameByAlpha  = "Алфавит"
	nameByYear   = "Год"
)

func composeMovieFileName(mov *model.Movie, f *dirEntry) string {
	_, fileName := path.Split(f.path)
	ext := path.Ext(f.path)

	switch mov.Info.Type {
	case rms_library.MovieType_Film:
		if f.info.EpisodeName == "" {
			return fileName
		}
		return escape(f.info.EpisodeName) + ext
	case rms_library.MovieType_TvSeries:
		if f.info.Episode < 0 {
			if f.info.EpisodeName == "" {
				return fileName
			}
			return f.info.EpisodeName + ext
		}
		if f.info.EpisodeName == "" {
			return fmt.Sprintf("E%02d%s", f.info.Episode, ext)
		}
		return fmt.Sprintf("E%02d. %s", f.info.Episode, fileName)
	case rms_library.MovieType_Clip:
		return f.relpath
	}

	return ""
}

func getMovieCategoryDir(mov *model.Movie) string {
	switch mov.Info.Type {
	case rms_library.MovieType_TvSeries:
		return nameTvSeries
	case rms_library.MovieType_Film:
		return nameFilms
	case rms_library.MovieType_Clip:
		return nameClips

	default:
		return ""
	}
}
