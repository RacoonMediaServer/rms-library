package model

import rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"

const MoviesCategory = "rms_movies"
const TvSeriesCategory = "rms_tv"
const ClipCategory = "rms_clip"

func GetCategory(t rms_library.MovieType) string {
	switch t {
	case rms_library.MovieType_Film:
		return MoviesCategory
	case rms_library.MovieType_TvSeries:
		return TvSeriesCategory
	case rms_library.MovieType_Clip:
		return ClipCategory
	}
	return ""
}
