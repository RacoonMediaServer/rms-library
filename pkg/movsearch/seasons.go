package movsearch

type Seasons map[uint]struct{}

func (s Seasons) Union(other Seasons) {
	for seasonNo := range other {
		s[seasonNo] = struct{}{}
	}
}
