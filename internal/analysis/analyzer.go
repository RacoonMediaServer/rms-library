package analysis

import (
	"regexp"
	"strconv"
)

type analyzeResult struct {
	Tokens  tokenList
	Year    uint
	Season  uint
	Episode int
}

type analyzeContext struct {
	result analyzeResult
	remove []bool
	name   tokenList
}

func analyzeFileName(name tokenList) analyzeResult {
	ctx := analyzeContext{
		name:   name,
		remove: make([]bool, len(name)),
	}
	ctx.result.Episode = -1

	// 0) отдельный code path для кейса, когда в названии сриала слитно указан сезон (например: StrangerThingsS03)
	determineNameSeasonCase(&ctx)

	// 1) пытаемся определить сезон, сначала считаем, что он указан отдельно (сезон 1, 1 сезон, season 1, etc)
	determineSplitSeason(&ctx)

	// 2) если не удалось ищем сочетания S01, season01, 01season и пр.
	determineSeason(&ctx)

	// 4) вытаскиваем год
	determineYear(&ctx)

	// 5) удаляем распознанные лишние слова
	removeExtraWords(&ctx)

	// 3) определяем эпизод
	determineEpisode(&ctx)

	// 6) пробуем угадать длину названия и обрезаем список
	titleLength := guessTitleLength(ctx.name, ctx.remove)
	ctx.result.Tokens = crop(ctx.name, ctx.remove, titleLength)

	return ctx.result
}

func determineNameSeasonCase(ctx *analyzeContext) {
	if len(ctx.name) != 1 {
		return
	}

	expr := regexp.MustCompile(`s\d\d?`)
	found := expr.FindString(ctx.name[0].Text)
	if found != "" {
		season, _ := strconv.ParseUint(found[1:], 10, 32)
		ctx.result.Season = uint(season)
		idx := expr.FindStringIndex(ctx.name[0].Text)
		tmp := ctx.name[0].Text
		tmp = tmp[:idx[0]] + tmp[idx[1]:]
		ctx.name[0].Text = tmp
	}
}

func determineSplitSeason(ctx *analyzeContext) {
	splitSeasonMatch := &orMatch{
		Matches: []match{
			&wordMatch{Word: "сезон"},
			&wordMatch{Word: "season"},
			&wordMatch{Word: "sezon"},
		},
	}
	pos := ctx.name.Find(splitSeasonMatch)
	if pos > -1 {
		m := regexMatch{Exp: regexp.MustCompile(`^\d\d?$`)}
		found := -1
		if pos < len(ctx.name)-1 && m.Match(ctx.name[pos+1]) {
			found = pos + 1
		} else if pos > 0 && m.Match(ctx.name[pos-1]) {
			found = pos - 1
		}
		if found < 0 && pos > 1 && ctx.name[pos-1].Text == "й" && m.Match(ctx.name[pos-2]) {
			found = pos - 2
		}

		if found > -1 {
			season, _ := strconv.ParseUint(ctx.name[found].Text, 10, 32)
			ctx.result.Season = uint(season)
			ctx.remove[pos] = true
			ctx.remove[found] = true
		}
	}
}

func determineSeason(ctx *analyzeContext) {
	seasonMatch := &orMatch{
		Matches: []match{
			&regexMatch{Exp: regexp.MustCompile(`s\d\d?`)},
			&regexMatch{Exp: regexp.MustCompile(`season\d\d?`)},
			&regexMatch{Exp: regexp.MustCompile(`сезон\d\d?`)},
			&regexMatch{Exp: regexp.MustCompile(`\d\d?season`)},
			&regexMatch{Exp: regexp.MustCompile(`сезон\d\d?`)},
		},
	}
	pos := ctx.name.Find(seasonMatch)
	if pos > -1 {
		exp := regexp.MustCompile(`\d\d?`)
		seasonString := exp.FindString(ctx.name[pos].Text)
		if seasonString == "" {
			return
		}
		season, _ := strconv.ParseUint(seasonString, 10, 32)
		ctx.result.Season = uint(season)
		ctx.remove[pos] = true
	}
}

func determineEpisode(ctx *analyzeContext) {
	rmatches := []regexMatch{
		{Exp: regexp.MustCompile(`e\d\d`)},
		{Exp: regexp.MustCompile(`x\d\d`)},
	}
	for _, m := range rmatches {
		pos := ctx.name.Find(m)
		if pos > -1 {
			episodeString := m.Exp.FindString(ctx.name[pos].Text)
			if episodeString == "" {
				continue
			}
			episode, _ := strconv.ParseInt(episodeString[1:], 10, 32)
			ctx.result.Episode = int(episode)
			ctx.remove[pos] = true

			return
		}
	}

	pos := 0
	m := &regexMatch{Exp: regexp.MustCompile(`^\d\d$`)}
	for {
		cpos := ctx.name[pos:].Find(m)
		if cpos < 0 {
			pos = -1
			break
		}
		pos = pos + cpos
		if !ctx.remove[pos] {
			break
		}
		pos++
	}

	if pos > -1 {
		episode, _ := strconv.ParseInt(ctx.name[pos].Text, 10, 32)
		ctx.result.Episode = int(episode)
		ctx.remove[pos] = true
		return
	}

	pos = ctx.name.Find(&regexMatch{Exp: regexp.MustCompile(`^\d$`)})
	if pos > -1 && !ctx.remove[pos] {
		episode, _ := strconv.ParseInt(ctx.name[pos].Text, 10, 32)
		ctx.result.Episode = int(episode)
	}
}

func determineYear(ctx *analyzeContext) {
	yearMatch := &regexMatch{Exp: regexp.MustCompile(`^\d\d\d\d$`)}
	pos := ctx.name.Find(yearMatch)
	if pos > -1 {
		year, _ := strconv.ParseUint(ctx.name[pos].Text, 10, 32)
		ctx.remove[pos] = true
		ctx.result.Year = uint(year)
	}
}

func removeExtraWords(ctx *analyzeContext) {
	matched := ctx.name.FindAll(
		&orMatch{
			Matches: []match{
				&bracesMatch{},
				&wordMatch{Word: "rus"},
				&wordMatch{Word: "eng"},
				&wordMatch{Word: "avo"},
				&wordMatch{Word: "remastered"},
				&wordMatch{Word: "web"},
				&wordMatch{Word: "dl"},
				&wordMatch{Word: "webdl"},
				&wordMatch{Word: "sub"},
				&wordMatch{Word: "lostfilm"},
				&wordMatch{Word: "unrated"},
				&wordMatch{Word: "dvd"},
				&wordMatch{Word: "сериал"},
				&wordMatch{Word: "серия"},
				&regexMatch{Exp: regexp.MustCompile(`rip$`)},
				&regexMatch{Exp: regexp.MustCompile(`^\d\d\d\d?p$`)},
				&regexMatch{Exp: regexp.MustCompile(`^\d\d\d\d$`)},
				&regexMatch{Exp: regexp.MustCompile(`remux$`)},
			},
		},
	)

	for _, r := range matched {
		ctx.remove[r] = true
	}
}

func guessTitleLength(name tokenList, remove []bool) int {
	for i, r := range remove {
		if r && i != 0 {
			if i == 1 && name[i-1].IsDigital() {
				continue
			}
			return i
		}
	}

	return len(remove)
}

func crop(name tokenList, remove []bool, maxLength int) tokenList {
	result := make([]token, 0, maxLength)
	for i, t := range name {
		if !remove[i] {
			result = append(result, t)
		}
		if len(result) >= maxLength {
			return result
		}
	}

	return result
}
