package analysis

import (
	"path/filepath"
	"strings"
)

type dirLayout struct {
	Primary   string
	Secondary string
	SubPath   string
	FileName  string
	Extension string
}

func extractLayout(file string) dirLayout {
	result := dirLayout{}

	result.Extension = filepath.Ext(file)
	if strings.HasPrefix(result.Extension, ".") {
		result.Extension = result.Extension[1:]
	}

	path, fileName := filepath.Split(file)
	result.FileName = strings.TrimSuffix(fileName, "."+result.Extension)

	directories := strings.Split(path, string(filepath.Separator))
	if len(directories) == 0 || directories[0] == "" {
		return result
	}

	result.Primary = directories[0]
	if len(directories) == 1 {
		return result
	}

	result.Secondary = directories[1]
	if len(directories) == 2 {
		return result
	}

	result.SubPath = filepath.Join(directories[2:]...)
	return result
}

func (l dirLayout) IsVideoFile() bool {

	var videoExtensions = []string{
		"mkv", "mp4", "vob", "sub", "3gp", "avi", "wmv", "flv", "ogv", "mp4v", "ts", "mpeg4", "mjpg", "mpg", "mov", "xvid",
	}

	ext := strings.ToLower(l.Extension)
	for _, videoExtension := range videoExtensions {
		if ext == videoExtension {
			return true
		}
	}

	return false
}

func (l dirLayout) IsSubtitlesFile() bool {

	var subtitleExtensions = []string{
		"srt", "vtt", "usf", "smil", "smi", "sami", "sub",
	}

	ext := strings.ToLower(l.Extension)
	for _, subtitleExtension := range subtitleExtensions {
		if ext == subtitleExtension {
			return true
		}
	}

	return false
}

func (l dirLayout) IsRootBased() bool {
	return len(l.Primary) == 0
}
