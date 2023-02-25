package model

import (
	"fmt"
	"path"
)

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

// String returns human-readable pretty formatted file name
func (f File) String() string {
	_, fileName := path.Split(f.Path)
	ext := path.Ext(f.Path)
	if f.No < 0 {
		if f.Title == "" {
			return fileName
		}
		return f.Title + ext
	}
	if f.Title == "" {
		return fmt.Sprintf("E%02d%s", f.No, ext)
	}
	return fmt.Sprintf("E%02d. %s", f.No, fileName)
}
