package main

import (
	"bufio"
	"context"
	"fmt"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"os"
	"strconv"
	"time"
)

const defaultTimeout = 60 * time.Second

func main() {
	var query string
	service := micro.NewService(
		micro.Name("rms-library.downloader"),
		micro.Flags(
			&cli.StringFlag{
				Name:        "query",
				Usage:       "Query for download",
				Required:    true,
				Destination: &query,
			},
		),
	)
	service.Init()

	library := rms_library.NewRmsLibraryService("rms-library", service.Client())

	results, err := library.SearchMovie(context.Background(), &rms_library.SearchMovieRequest{Text: query, Limit: 5}, client.WithRequestTimeout(defaultTimeout))
	if err != nil {
		panic(err)
	}

	for i, m := range results.Movies {
		if m.Info.Seasons != nil {
			fmt.Printf("#%d. %s [seasons %+v]\n", i+1, m.Info.Title, *m.Info.Seasons)
		} else {
			fmt.Printf("#%d. %s\n", i+1, m.Info.Title)
		}
	}
	fmt.Println("\nSelect which one:")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	no, err := strconv.ParseInt(scanner.Text(), 10, 32)
	if err != nil {
		panic(err)
	}

	resp, err := library.DownloadMovie(context.Background(), &rms_library.DownloadMovieRequest{Id: results.Movies[no-1].Id}, client.WithRequestTimeout(defaultTimeout))
	if err != nil {
		panic(err)
	}
	fmt.Println("Found: ", resp.Found)
	fmt.Println("Seasons: ", resp.Seasons)
}
