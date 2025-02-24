package movsearch

type Result struct {
	Torrent []byte
	Seasons Seasons
}

func GetMultipleResultsSeasons(results []Result) Seasons {
	result := Seasons{}
	for _, t := range results {
		for no := range t.Seasons {
			result[no] = struct{}{}
		}
	}
	return result
}
