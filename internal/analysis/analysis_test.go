package analysis

import (
	"testing"

	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"

	"github.com/stretchr/testify/assert"
)

func TestAnalyze(t *testing.T) {
	type testCase struct {
		input  string
		output Result
	}

	testCases := []testCase{
		{
			input: "Паранормальный Веллингтон. Сериал. Ozz (HDTVRip 720p)/2 Сезон (2019)/s02e03. Гудок в тоннеле Виктории.mkv",
			output: Result{
				Titles:      []string{"Паранормальный Веллингтон", "Гудок в Тоннеле Виктории"},
				Year:        2019,
				Season:      2,
				Episode:     3,
				BelongsTo:   rms_library.MovieType_TvSeries,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "Гудок в Тоннеле Виктории",
			},
		},
		{
			input: "Мир дикого запада (Сезон 1-4) Amedia/Сезон 2/Westworld.S02E05.BDRip.RGzsRutracker.Celta88.avi",
			output: Result{
				Titles:      []string{"Мир Дикого Запада", "Westworld"},
				Season:      2,
				Episode:     5,
				BelongsTo:   rms_library.MovieType_TvSeries,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "Westworld",
			},
		},
		{
			input: "Ugly.Americans.Season.1-2.2010-2012.x264.WEB-DL.720p.Zuich32/2 Season/05 The Ring of Powers.mkv",
			output: Result{
				Titles:      []string{"Ugly Americans", "The Ring of Powers"},
				Season:      2,
				Episode:     5,
				BelongsTo:   rms_library.MovieType_TvSeries,
				Year:        2010,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "The Ring of Powers",
			},
		},
		{
			input: "The_Guild.S01-06.720p.rus.stopgame.ru/The_Guild.S06.720p.rus.stopgame.ru/Гильдия 6-й сезон. Эпизод 12 Завершение игры Игровое кино - .mp4",
			output: Result{
				Titles:      []string{"The Guild", "Гильдия"},
				Season:      6,
				Episode:     12,
				BelongsTo:   rms_library.MovieType_TvSeries,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "Гильдия",
			},
		},
		{
			input: "StarGate SG-1/SG-1. Season-10/SG-1. Season 10.02. Morpheus.avi",
			output: Result{
				Titles:      []string{"Stargate sg 1", "sg 1"},
				Season:      10,
				Episode:     2,
				BelongsTo:   rms_library.MovieType_TvSeries,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "sg 1",
			},
		},
		{
			input: "StarGate SG-1/SG-1. Season-10/SG-1. Season 10.02. Morpheus.srt",
			output: Result{
				Titles:      []string{"Stargate sg 1", "sg 1"},
				BelongsTo:   rms_library.MovieType_TvSeries,
				FileType:    model.FileTypeMediaSupply,
				Season:      10,
				Episode:     2,
				EpisodeName: "sg 1",
			},
		},
		{
			input: "Disenchantment.2018.web-dlrip_[teko]/Season_02/s02e07_Bad.Moon.Rising.avi",
			output: Result{
				Titles:      []string{"Disenchantment", "Bad Moon Rising"},
				Season:      2,
				Episode:     7,
				Year:        2018,
				BelongsTo:   rms_library.MovieType_TvSeries,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "Bad Moon Rising",
			},
		},
		{
			input: "Хан Соло Звёздные Войны. Истории.2018.UHD.BDRip.1080p.HDR.mkv",
			output: Result{
				Titles:      []string{"Хан Соло Звёздные Войны Истории"},
				Episode:     -1,
				Year:        2018,
				BelongsTo:   rms_library.MovieType_Film,
				FileType:    model.FileTypeFilm,
				EpisodeName: "Хан Соло Звёздные Войны Истории",
			},
		},
		{
			input: "Стражи Галактики_2.1080p. Ton.mkv",
			output: Result{
				Titles:      []string{"Стражи Галактики 2"},
				Episode:     2,
				BelongsTo:   rms_library.MovieType_Film,
				FileType:    model.FileTypeFilm,
				EpisodeName: "Стражи Галактики 2",
			},
		},
		{
			input: "Полицейский с Рублёвки. Снова дома WEB-DL 1080p (Версия без цензуры)/02 серия.mkv",
			output: Result{
				Titles:  []string{"Полицейский с Рублёвки Снова Дома"},
				Episode: 2,
				// TODO: придумать, что делать в таком случае
				BelongsTo: rms_library.MovieType_TvSeries,
				FileType:  model.FileTypeEpisode,
			},
		},
		{
			input: "[SOFCJ-Raws] X-Men TAS The Complete Series [DVDRip]/[SOFCJ-Raws] X-Men TAS 1x07 (07) Slave Island [DVDRip].mkv",
			output: Result{
				Titles:      []string{"x Men Tas The Complete Series", "x Men Tas"},
				BelongsTo:   rms_library.MovieType_TvSeries,
				Episode:     7,
				Season:      1,
				FileType:    model.FileTypeEpisode,
				EpisodeName: "x Men Tas",
			},
		},
	}

	for i, tc := range testCases {
		assert.Equal(t, tc.output, Analyze(tc.input), "Test %d failed", i)
	}
}
