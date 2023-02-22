package storage

func getFirst[K comparable, V any](m map[K]V) (K, V) {
	for k, v := range m {
		return k, v
	}
	panic("map must be not empty")
}
