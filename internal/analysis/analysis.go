package analysis

import (
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
)

type Result struct {
	Titles      []string
	Year        uint
	BelongsTo   rms_library.MovieType
	FileType    model.FileType
	Season      uint
	Episode     int
	EpisodeName string
}

func Analyze(fileName string) Result {
	layout := extractLayout(fileName)

	subResults := analyzeLayout(layout)
	result := mergeResults(subResults)
	if layout.IsVideoFile() {
		if result.Episode != 0 {
			result.FileType = model.FileTypeEpisode
		} else {
			result.FileType = model.FileTypeFilm
		}
	}
	if layout.IsSubtitlesFile() {
		result.FileType = model.FileTypeMediaSupply
	}
	return result
}

func analyzeLayout(layout dirLayout) []analyzeResult {
	var results []analyzeResult
	if layout.Primary != "" {
		results = append(results, analyzeFileName(parseName(layout.Primary)))
	}
	if layout.Secondary != "" {
		results = append(results, analyzeFileName(parseName(layout.Secondary)))
	}
	if layout.SubPath != "" {
		results = append(results, analyzeFileName(parseName(layout.SubPath)))
	}
	if layout.FileName != "" {
		results = append(results, analyzeFileName(parseName(layout.FileName)))
	}
	return results
}

func mergeResults(results []analyzeResult) Result {
	result := Result{}

	// определяем название эпизода
	if len(results) != 0 {
		result.EpisodeName = results[len(results)-1].Tokens.String()
	}
	// определяем епизод
	result.Episode = -1
	for i := len(results) - 1; i >= 0; i-- {
		if result.Episode < 0 {
			result.Episode = results[i].Episode
		}
	}

	// определяем сезон
	for i := len(results) - 1; i >= 0; i-- {
		if result.Season == 0 {
			result.Season = results[i].Season
		}
	}

	// определяем год
	for i := range results {
		if result.Year == 0 {
			result.Year = results[i].Year
		}
	}

	// заполняем названия
	titles := map[string]bool{}
	for i := range results {
		title := results[i].Tokens.String()
		_, exist := titles[title]
		if title != "" && !exist {
			result.Titles = append(result.Titles, title)
			titles[title] = true
		}
	}

	result.BelongsTo = rms_library.MovieType_Film
	if result.Season != 0 {
		result.BelongsTo = rms_library.MovieType_TvSeries
	}

	return result
}
