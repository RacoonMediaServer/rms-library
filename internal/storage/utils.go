package storage

import (
	"strings"
	"unicode"
)

func getFirst[K comparable, V any](m map[K]V) (K, V) {
	for k, v := range m {
		return k, v
	}
	panic("map must be not empty")
}

func escape(s string) string {
	return strings.Replace(s, "/", "", -1)
}

func capitalize(s string) string {
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
