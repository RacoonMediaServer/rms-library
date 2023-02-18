package service

import (
	"context"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/go-openapi/runtime"
	"google.golang.org/protobuf/types/known/emptypb"
)

const searchTorrentsLimit uint = 10

type LibraryService struct {
	c    *cache
	f    servicemgr.ServiceFactory
	auth runtime.ClientAuthInfoWriter
	db   Database
	cli  *client.Client
}

func (l LibraryService) SearchMovie(ctx context.Context, request *rms_library.SearchMovieRequest, response *rms_library.SearchMovieResponse) error {
	//TODO implement me
	panic("implement me")
}

func (l LibraryService) DownloadMovie(ctx context.Context, request *rms_library.DownloadMovieRequest, response *rms_library.DownloadMovieResponse) error {
	//TODO implement me
	panic("implement me")
}

func (l LibraryService) GetTvSeriesUpdates(ctx context.Context, empty *emptypb.Empty, response *rms_library.GetTvSeriesUpdatesResponse) error {
	//TODO implement me
	panic("implement me")
}

func (l LibraryService) GetMovieInfo(ctx context.Context, request *rms_library.GetMovieInfoRequest, info *rms_library.MovieInfo) error {
	//TODO implement me
	panic("implement me")
}

func (l LibraryService) GetMovies(ctx context.Context, request *rms_library.GetMoviesRequest, response *rms_library.GetMoviesResponse) error {
	//TODO implement me
	panic("implement me")
}

func (l LibraryService) GetMovieLayout(ctx context.Context, request *rms_library.GetMovieLayoutRequest, response *rms_library.GetMovieLayoutResponse) error {
	//TODO implement me
	panic("implement me")
}

func NewService(db Database, f servicemgr.ServiceFactory, cli *client.Client, auth runtime.ClientAuthInfoWriter) rms_library.RmsLibraryHandler {
	return &LibraryService{
		c:    newCache(),
		f:    f,
		auth: auth,
		db:   db,
		cli:  cli,
	}
}
