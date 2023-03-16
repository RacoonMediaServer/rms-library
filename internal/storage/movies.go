package storage

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"path"
	"unicode"
	"unicode/utf8"
)

func getMovieDirectories(mov *model.Movie) (directories []string) {
	title := escape(mov.Info.Title)

	directories = append(directories, path.Join(getCategory(mov), title))

	if mov.Info.Year != 0 {
		directories = append(directories, path.Join(nameByYear, fmt.Sprintf("%d", mov.Info.Year), title))
	}

	letter, _ := utf8.DecodeRuneInString(title)
	letter = unicode.ToUpper(letter)
	directories = append(directories, path.Join(nameByAlpha, string(letter), title))

	for _, g := range mov.Info.Genres {
		g = capitalize(g)
		directories = append(directories, path.Join(nameByGenre, g, title))
	}

	return
}
