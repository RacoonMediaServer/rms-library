package movsearch

import (
	"hash/crc32"

	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
)

func extractSeasonsFromResult(result *models.SearchTorrentsResult) Seasons {
	seasons := Seasons{}
	for _, seasonNo := range result.Seasons {
		seasons[uint(seasonNo)] = struct{}{}
	}
	return seasons
}

func removeDuplicatedResults(src []Result) []Result {
	if len(src) < 2 {
		return src
	}

	crc := map[uint32]Result{}
	for _, r := range src {
		crc[crc32.ChecksumIEEE(r.Torrent)] = r
	}

	result := make([]Result, 0, len(crc))
	for _, r := range crc {
		result = append(result, r)
	}

	return result
}
