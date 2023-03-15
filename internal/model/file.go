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

// File represents media entry
type File struct {
	// Relative Path on data storage
	Path string

	// Type defines significance of File
	Type FileType

	// Title is a human-readable name of episode/track other
	Title string

	// No means episode/track number, can be -1
	No int
}
