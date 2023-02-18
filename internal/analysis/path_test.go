package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractLayout(t *testing.T) {
	type testCase struct {
		input  string
		output dirLayout
	}

	testCases := []testCase{
		{
			input: "somefile",
			output: dirLayout{
				FileName: "somefile",
			},
		},
		{
			input: "movie.mp4",
			output: dirLayout{
				FileName:  "movie",
				Extension: "mp4",
			},
		},
		{
			input: "serial/movie",
			output: dirLayout{
				Primary:  "serial",
				FileName: "movie",
			},
		},
		{
			input: "serial/movie.mp4",
			output: dirLayout{
				Primary:   "serial",
				FileName:  "movie",
				Extension: "mp4",
			},
		},
		{
			input: "serial/season 1/movie.mp4",
			output: dirLayout{
				Primary:   "serial",
				Secondary: "season 1",
				FileName:  "movie",
				Extension: "mp4",
			},
		},
		{
			input: "serial/season 1/subdir/movie.mp4",
			output: dirLayout{
				Primary:   "serial",
				Secondary: "season 1",
				SubPath:   "subdir",
				FileName:  "movie",
				Extension: "mp4",
			},
		},
		{
			input: "serial/season 1/subdir1/subdir2/movie.mp4",
			output: dirLayout{
				Primary:   "serial",
				Secondary: "season 1",
				SubPath:   "subdir1/subdir2",
				FileName:  "movie",
				Extension: "mp4",
			},
		},
		{
			input: "serial/season 1/subdir1/subdir2/subdir3/movie.mp4",
			output: dirLayout{
				Primary:   "serial",
				Secondary: "season 1",
				SubPath:   "subdir1/subdir2/subdir3",
				FileName:  "movie",
				Extension: "mp4",
			},
		},
	}

	for i, tc := range testCases {
		actual := extractLayout(tc.input)
		assert.Equal(t, tc.output, actual, "Test %d failed", i)
	}
}
