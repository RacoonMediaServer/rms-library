package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeFileName(t *testing.T) {
	type testCase struct {
		input  string
		output analyzeResult
	}

	testCases := []testCase{
		{
			input: "SG-1. Season 10.03. The Pegasus Project",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "sg"},
					token{Text: "1"},
				},
				Season:  10,
				Episode: 3,
			},
		},
		{
			input: "Stranger.Things.S04.WEBDL.1080p.Rus.Eng",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "stranger"},
					token{Text: "things"},
				},
				Season:  4,
				Episode: -1,
			},
		},
		{
			input: "Lexx.1997-2001.dvdrip_[teko]",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "lexx"},
				},
				Year:    1997,
				Episode: -1,
			},
		},
		{
			input: "Babylon.5.1993-2007.dvdrip_[full.collection]_[teko]",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "babylon"},
					token{Text: "5"},
				},
				Year:    1993,
				Episode: 5,
			},
		},
		{
			input: "Altered.Carbon.S01.1080p.NF.WEBRip.4xRus.Eng.sergiy_psp",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "altered"},
					token{Text: "carbon"},
				},
				Season:  1,
				Episode: -1,
			},
		},
		{
			input: "Highlander.1986.REMASTERED.1080p.BluRay.16xRus.2xUkr.2xEng.TeamHD-Атлас31",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "highlander"},
				},
				Year:    1986,
				Episode: -1,
			},
		},
		{
			input: "Assasin_Bitva_mirov_WEB-DLRip_by_Dalemake",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "assasin"},
					token{Text: "bitva"},
					token{Text: "mirov"},
				},
				Episode: -1,
			},
		},
		{
			input: "Brassic",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "brassic"},
				},
				Episode: -1,
			},
		},
		{
			input: "Season 01",
			output: analyzeResult{
				Tokens:  tokenList{},
				Season:  1,
				Episode: -1,
			},
		},
		{
			input: "s01e01_Pilot",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "pilot"},
				},
				Season:  1,
				Episode: 1,
			},
		},
		{
			input: "Sejlor.Mun.S.03.serija.iz.38.avi",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "sejlor"},
					token{Text: "mun"},
					token{Text: "s"},
				},
				Episode: 3,
			},
		},
		{
			input: "The.Owl.House.S01E18.Agony.of.a.Witch.1080p.WEB-DL.RU.Rus.Eng_WORTEXSON",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "the"},
					token{Text: "owl"},
					token{Text: "house"},
				},
				Season:  1,
				Episode: 18,
			},
		},
		{
			input: "15 The Stalking Dead",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "the"},
					token{Text: "stalking"},
					token{Text: "dead"},
				},
				Episode: 15,
			},
		},
		{
			input: "1-04 The Box (HD)",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "1"},
					token{Text: "the"},
					token{Text: "box"},
				},
				Episode: 4,
			},
		},
		{
			input: "Гильдия 6-й сезон. Эпизод 12 Завершение игры Игровое кино - ",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "гильдия"},
				},
				Season:  6,
				Episode: 12,
			},
		},
		{
			input: "12.Monkeys.S02E04.HDRip.RGzsRutracker.Celta88.avi",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "12"},
					token{Text: "monkeys"},
				},
				Season:  2,
				Episode: 4,
			},
		},
		{
			input: "Эш против Зловещих мертвецов. 2 Сезон. 2016 (Blu-Ray Remux 1080p)/Ash.vs.Evil.Dead.s02e08. Эш — всех зарежь.",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "эш"},
					token{Text: "против"},
					token{Text: "зловещих"},
					token{Text: "мертвецов"},
				},
				Season:  2,
				Episode: 8,
				Year:    2016,
			},
		},
		{
			input: "Ash vs Evil Dead (Season 01) 1080p",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "ash"},
					token{Text: "vs"},
					token{Text: "evil"},
					token{Text: "dead"},
				},
				Season:  1,
				Episode: -1,
			},
		},
		{
			input: "Паранормальный Веллингтон. 3 сезон. 2021. EniaHD (HDTV 1080p)",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "паранормальный"},
					token{Text: "веллингтон"},
				},
				Season:  3,
				Year:    2021,
				Episode: -1,
			},
		},
		{
			input: "StrangerThingsS03",
			output: analyzeResult{
				Tokens: tokenList{
					token{Text: "strangerthings"},
				},
				Season:  3,
				Episode: -1,
			},
		},
	}

	for i, tc := range testCases {
		tokens := parseName(tc.input)
		actual := analyzeFileName(tokens)
		assert.Equal(t, tc.output, actual, "Test %d failed", i)
	}
}
