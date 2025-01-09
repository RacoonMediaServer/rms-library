package service

import (
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/pkg/selector"
)

func (l LibraryService) getMovieSelector(mov *model.Movie) selector.MediaSelector {
	// TODO: вынести в настройки

	settings := selector.Settings{
		MinSeasonSizeMB:     1024,
		MaxSeasonSizeMB:     50 * 1024,
		MinSeedersThreshold: 50,
		QualityPrior:        []string{"1080p", "720p", "480p"},
		Voice:               mov.Voice,
	}

	settings.VoiceList.Append("сыендук", "syenduk")
	settings.VoiceList.Append("кубик", "кубе", "kubik", "kube")
	settings.VoiceList.Append("кураж", "бомбей", "kurazh", "bombej")
	settings.VoiceList.Append("lostfilm", "lost")
	settings.VoiceList.Append("newstudio")
	settings.VoiceList.Append("амедиа", "amedia")

	return selector.New(settings)
}
