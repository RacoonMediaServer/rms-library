package storage

import (
	"fmt"
	"path"

	"github.com/RacoonMediaServer/rms-library/v3/internal/analysis"
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

func composeMovieFileName(mi *rms_library.MovieInfo, fullPath string, info *analysis.Result) string {
	_, fileName := path.Split(fullPath)
	ext := path.Ext(fullPath)

	switch mi.Type {
	case rms_library.MovieType_Film:
		if info.EpisodeName == "" {
			return fileName
		}
		return escape(info.EpisodeName) + ext
	case rms_library.MovieType_TvSeries:
		if info.Episode < 0 {
			if info.EpisodeName == "" {
				return fileName
			}
			return info.EpisodeName + ext
		}
		if info.EpisodeName == "" {
			return fmt.Sprintf("E%02d%s", info.Episode, ext)
		}
		return fmt.Sprintf("E%02d. %s", info.Episode, fileName)
	}

	return ""
}

func getMovieCategoryDir(mi *rms_library.MovieInfo) string {
	switch mi.Type {
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
