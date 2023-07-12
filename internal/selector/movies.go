package selector

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/antzucaro/matchr"
	"go-micro.dev/v4/logger"
	"strings"
)

type MovieSelector struct {
	MinSeasonSizeMB     int64
	MaxSeasonSizeMB     int64
	MinSeedersThreshold int64
	Voice               string
	VoiceList           Voices
	QualityPrior        []string
}

func (s MovieSelector) getRankFunction(criteria Criteria) rankFunc {
	switch criteria {
	case CriteriaQuality:
		return makeRankFunc(s.limitBySize, s.rankByQuality, s.getRankByVoiceFunc(2))
	case CriteriaFastest:
		return makeRankFunc(s.rankBySize, s.rankBySeeders, s.getRankByVoiceFunc(0.5))
	case CriteriaCompact:
		return makeRankFunc(s.limitBySize, s.rankBySeeders, s.rankByQuality, s.rankBySeasons, s.getRankByVoiceFunc(2))
	}
	panic("unknown criteria")
}

func (s MovieSelector) Select(criteria Criteria, list []*models.SearchTorrentsResult) *models.SearchTorrentsResult {
	rank := s.getRankFunction(criteria)
	ranks := rank(list)
	_, _, best := findMax(ranks, func(elem float32) float32 {
		return elem
	})
	for i := range ranks {
		logger.Tracef("%d rank: %.4f", i, ranks[i])
	}
	sel := list[best]
	logger.Infof("Selected { Title: %s, Voice: %s, Size: %d, Seeders: %d, Quality: %s }", getString(sel.Title), sel.Voice, getValue(sel.Size), getValue(sel.Seeders), sel.Quality)
	return sel
}

func (s MovieSelector) rankBySize(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	_, max, _ := findMax(list, func(t *models.SearchTorrentsResult) int64 {
		return getValue(t.Size)
	})

	for i, t := range list {
		ranks[i] = 1 - (float32(getValue(t.Size)) / float32(max))
		logger.Tracef("%d rank by size: %.4f", i, ranks[i])
	}
	return ranks
}

func (s MovieSelector) limitBySize(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))

	for i, t := range list {
		size := getValue(t.Size)
		seasons := int64(len(t.Seasons))
		if seasons == 0 {
			seasons = 1
		}

		if size < s.MinSeasonSizeMB*seasons || size >= s.MaxSeasonSizeMB*seasons {
			ranks[i] = -1
			logger.Tracef("%d limit by size: %.4f", i, ranks[i])
		}

	}
	return ranks
}

func (s MovieSelector) rankBySeeders(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	_, max, _ := findMax(list, func(t *models.SearchTorrentsResult) int64 {
		return getValue(t.Seeders)
	})

	for i, t := range list {
		seeders := getValue(t.Seeders)
		if seeders < s.MinSeedersThreshold {
			ranks[i] = float32(seeders) / float32(max)
		} else {
			ranks[i] = 1
		}
		logger.Tracef("%d rank by seeders: %.4f", i, ranks[i])
	}
	return ranks
}

func (s MovieSelector) rankByQuality(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	perQualityWeight := 1 / float32(len(s.QualityPrior))
	for i, t := range list {
		for j, q := range s.QualityPrior {
			if t.Quality == q {
				ranks[i] = float32(len(s.QualityPrior)-j) * perQualityWeight
				break
			}
		}
		logger.Tracef("%d rank by quality: %.4f", i, ranks[i])
	}

	return ranks
}

func (s MovieSelector) rankBySeasons(list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	_, max, _ := findMax(list, func(t *models.SearchTorrentsResult) int {
		return len(t.Seasons)
	})

	for i, t := range list {
		ranks[i] = float32(len(t.Seasons)) / float32(max)
	}

	return ranks
}

func (s MovieSelector) getRankByVoiceFunc(weight float32) rankFunc {
	return func(list []*models.SearchTorrentsResult) []float32 {
		if s.Voice == "" {
			return s.rankByVoiceList(weight, list)
		}
		return s.rankByVoice(weight, list)
	}
}

func (s MovieSelector) rankByVoiceList(weight float32, list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	perItemWeight := 1 / float32(len(s.VoiceList))
	for i, t := range list {
		tVoice := strings.ToLower(t.Voice)
	ScanVoice:
		for j, voice := range s.VoiceList {
			for _, w := range voice {
				if strings.Index(tVoice, w) >= 0 {
					ranks[i] = weight * float32(len(s.VoiceList)-j) * perItemWeight
					logger.Tracef("%d rank by voice list: %.4f", i, ranks[i])
					break ScanVoice
				}
			}
		}
	}
	return ranks
}

func (s MovieSelector) rankByVoice(weight float32, list []*models.SearchTorrentsResult) []float32 {
	ranks := make([]float32, len(list))
	distance := make([]int, len(list))

	target := strings.ToLower(s.Voice)
	for i, t := range list {
		voice := strings.ToLower(t.Voice)
		distance[i] = matchr.Levenshtein(voice, target)
	}

	_, max, _ := findMax(distance, func(elem int) int {
		return elem
	})
	for j, d := range distance {
		ranks[j] = weight * (1 - (float32(d) / float32(max)))
		logger.Tracef("%d rank by voice: %.4f", j, ranks[j])
	}
	return ranks
}
