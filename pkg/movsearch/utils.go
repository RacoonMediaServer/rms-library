package movsearch

import "github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"

func extractSeasonsFromResult(result *models.SearchTorrentsResult) Seasons {
	seasons := Seasons{}
	for _, seasonNo := range result.Seasons {
		seasons[uint(seasonNo)] = struct{}{}
	}
	return seasons
}
