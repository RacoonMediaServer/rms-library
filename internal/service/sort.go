package service

import (
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"go-micro.dev/v4/logger"
	"sort"
)

func sortTorrentMoviesByQuality(list []*models.SearchTorrentsResult) {
	// хотим в приоритете иметь 1080p, в дальнейшем следует вынести в настройки
	qualityPrior := map[string]int{
		"1080p": 4,
		"720p":  3,
		"480p":  2,
		"":      1,
	}
	sort.SliceStable(list, func(i, j int) bool {
		return qualityPrior[list[i].Quality] > qualityPrior[list[j].Quality]
	})
}

func sortTorrentMoviesBySeasons(list []*models.SearchTorrentsResult) {
	sort.SliceStable(list, func(i, j int) bool {
		return len(list[i].Seasons) > len(list[j].Seasons)
	})
}

func sortTorrentMoviesByFast(list []*models.SearchTorrentsResult) {
	const requiredSeedersCount = 10

	// отсортировали по размеру, но нужно еще учитывать кол-во сидов
	sort.SliceStable(list, func(i, j int) bool {
		return *list[i].Size < *list[j].Size
	})

searchSuitable:
	for limit := requiredSeedersCount; limit > 0; limit-- {
		// ищем первую с начала списка раздачу, у которой кол-во сидов максимально близко к вменяемому
		for i, t := range list {
			if *t.Seeders >= int64(limit) {
				list[0], list[i] = list[i], list[0]
				break searchSuitable
			}
		}
	}

	logger.Infof("Selected faster torrent: size = %d, seeders = %d", *list[0].Size, *list[0].Seeders)
}

func selectSuitableTorrent(list []*models.SearchTorrentsResult) *models.SearchTorrentsResult {
	const maxTorrentSizePerSeasonMB = 70 * 1024

	for _, t := range list {
		if t.Size != nil && *t.Size >= maxTorrentSizePerSeasonMB && len(t.Seasons) < 2 {
			continue
		}
		return t
	}

	return list[0]
}
