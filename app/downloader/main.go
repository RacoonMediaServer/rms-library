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
)

const defaultTimeout = 5 * time.Minute

func main() {
	var query string
	var command string
	var list uint
	var torrentId string
	service := micro.NewService(
		micro.Name("rms-library.downloader"),
		micro.Flags(
			&cli.StringFlag{
				Name:        "command",
				Usage:       "add,list,delete,move,torrents-list,torrents-delete,torrents-find,torrent-replace",
				Required:    true,
				Destination: &command,
			},
			&cli.StringFlag{
				Name:        "query",
				Usage:       "Query for download or media id",
				Required:    false,
				Destination: &query,
			},
			&cli.StringFlag{
				Name:        "tid",
				Usage:       "Torrent ID",
				Required:    false,
				Destination: &torrentId,
			},
			&cli.UintFlag{
				Name:        "list",
				Usage:       "List (0 - favourite, 1 - watch list, 2 - archive)",
				Required:    false,
				Destination: &list,
			},
		),
	)
	service.Init()

	switch command {
	case "add":
		addCommand(service.Client(), query, rms_library.List(list))
	case "list":
		listCommand(service.Client(), rms_library.List(list))
	case "delete":
		deleteCommand(service.Client(), query)
	case "move":
		moveCommand(service.Client(), query, rms_library.List(list))
	case "torrents-list":
		torrentsListCommand(service.Client(), query)
	case "torrents-delete":
		torrentsDeleteCommand(service.Client(), query, torrentId)
	case "torrents-find":
		torrentsFindCommand(service.Client(), query)
	case "torrents-replace":
		torrentsReplaceCommand(service.Client(), query, torrentId)
	default:
		panic("unknown command")
	}

}

func addCommand(cli client.Client, query string, list rms_library.List) {
	library := rms_library.NewMoviesService("rms-library", cli)
	lists := rms_library.NewListsService("rms-library", cli)

	results, err := library.Search(context.Background(), &rms_library.MoviesSearchRequest{Text: query, Limit: 5}, client.WithRequestTimeout(defaultTimeout))
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

	_, err = lists.Add(context.Background(), &rms_library.ListsAddRequest{Id: results.Movies[no-1].Id, List: list}, client.WithRequestTimeout(defaultTimeout))
	if err != nil {
		panic(err)
	}
}

func listCommand(cli client.Client, list rms_library.List) {
	lists := rms_library.NewListsService("rms-library", cli)
	items, err := lists.List(context.Background(), &rms_library.ListsListRequest{List: list})
	if err != nil {
		panic(err)
	}

	for _, item := range items.Items {
		fmt.Printf("%s [ %s ] %d Mb \n", item.Title, item.Id, item.Size)
	}
}

func deleteCommand(cli client.Client, id string) {
	lists := rms_library.NewListsService("rms-library", cli)
	_, err := lists.Delete(context.Background(), &rms_library.ListsDeleteRequest{Id: id})
	if err != nil {
		panic(err)
	}
}

func moveCommand(cli client.Client, id string, list rms_library.List) {
	lists := rms_library.NewListsService("rms-library", cli)

	_, err := lists.Move(context.Background(), &rms_library.ListsMoveRequest{Id: id, List: list})
	if err != nil {
		panic(err)
	}
}

func torrentsListCommand(cli client.Client, id string) {
	torrents := rms_library.NewTorrentsService("rms-library", cli)
	list, err := torrents.List(context.Background(), &rms_library.TorrentsListRequest{Id: id})
	if err != nil {
		panic(err)
	}

	for _, t := range list.Torrents {
		fmt.Printf("%s [ %s ] %d Mb\n", t.Title, t.Id, t.Size)
	}
}

func torrentsDeleteCommand(cli client.Client, id, tId string) {
	torrents := rms_library.NewTorrentsService("rms-library", cli)
	_, err := torrents.Delete(context.Background(), &rms_library.TorrentsDeleteRequest{Id: id, TorrentId: tId})
	if err != nil {
		panic(err)
	}
}

func torrentsFindCommand(cli client.Client, id string) {
	torrents := rms_library.NewTorrentsService("rms-library", cli)
	list, err := torrents.FindAlternatives(context.Background(), &rms_library.TorrentsFindAlternativesRequest{Id: id})
	if err != nil {
		panic(err)
	}

	for i, t := range list.Torrents {
		fmt.Printf("%d. %s [ %s ] seeders:%d, %d Mb\n", i+1, t.Title, t.Id, t.Seeders, t.Size)
	}
}

func torrentsReplaceCommand(cli client.Client, id, tId string) {
	torrents := rms_library.NewTorrentsService("rms-library", cli)

	list, err := torrents.List(context.Background(), &rms_library.TorrentsListRequest{Id: id})
	if err != nil {
		panic(err)
	}

	if len(list.Torrents) == 0 {
		_, err = torrents.Add(context.Background(), &rms_library.TorrentsAddRequest{
			Id:           id,
			NewTorrentId: tId,
		})
		if err != nil {
			panic(err)
		}
		return
	}

	_, err = torrents.Replace(context.Background(), &rms_library.TorrentsReplaceRequest{Id: id, TorrentId: list.Torrents[0].Id, NewTorrentId: &tId})
	if err != nil {
		panic(err)
	}
}
