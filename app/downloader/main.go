package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/client"
	"go-micro.dev/v4/logger"
)

const defaultTimeout = 2 * time.Minute

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

	library := rms_library.NewMoviesService("rms-library", service.Client())

	watchList, err := library.GetWatchList(context.Background(), &rms_library.GetMoviesRequest{}, client.WithRequestTimeout(defaultTimeout))
	if err != nil {
		panic(err)
	}

	for _, mov := range watchList.Result {
		logger.Infof("Item '%s' found in the watchlist", mov.Info.Title)
	}

	results, err := library.Search(context.Background(), &rms_library.SearchRequest{Text: query, Limit: 5}, client.WithRequestTimeout(defaultTimeout))
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

	resp, err := library.DownloadAuto(context.Background(), &rms_library.DownloadMovieAutoRequest{Id: results.Movies[no-1].Id, UseWatchList: true}, client.WithRequestTimeout(defaultTimeout))
	if err != nil {
		panic(err)
	}
	fmt.Println("Found: ", resp.Found)
	fmt.Println("Seasons: ", resp.Seasons)
	// _, err = library.WatchLater(context.Background(), &rms_library.WatchLaterRequest{Id: results.Movies[no-1].Id}, client.WithRequestTimeout(defaultTimeout))
	// if err != nil {
	// 	panic(err)
	// }
}
