package service

import (
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/internal/selector"
)

func (l LibraryService) getMovieSelector(mov *model.Movie) selector.MovieSelector {
	// TODO: вынести в настройки
	sel := selector.MovieSelector{
		MinSeasonSizeMB:     1024,
		MaxSeasonSizeMB:     50 * 1024,
		MinSeedersThreshold: 50,
		QualityPrior:        []string{"1080p", "720p", "480p"},
		Voice:               mov.Voice,
	}

	sel.VoiceList.Append("сыендук", "syenduk")
	sel.VoiceList.Append("кубик", "кубе", "kubik", "kube")
	sel.VoiceList.Append("кураж", "бомбей", "kurazh", "bombej")
	sel.VoiceList.Append("lostfilm", "lost")
	sel.VoiceList.Append("newstudio")
	sel.VoiceList.Append("амедиа", "amedia")

	return sel
}
