package model

type FileType int

const (
	// FileTypeInsignificant means regular trash files (readme, text, images, etc)
	FileTypeInsignificant FileType = iota

	// FileTypeFilm for films
	FileTypeFilm

	// FileTypeEpisode means episode of TV series
	FileTypeEpisode

	// FileTypeMediaSupply means subtitles, audio tracks and other
	FileTypeMediaSupply
)
