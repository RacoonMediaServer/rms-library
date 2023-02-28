package service

import (
	"context"
	"github.com/RacoonMediaServer/rms-library/internal/config"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"google.golang.org/protobuf/types/known/emptypb"
)

const searchTorrentsLimit uint = 10

// LibraryService is a service API handler
type LibraryService struct {
	f                servicemgr.ServiceFactory
	auth             runtime.ClientAuthInfoWriter
	db               Database
	cli              *client.Client
	dir              DirectoryManager
	dm               DownloadsManager
	torrentToMovieID map[string]string
}

// Settings holds all dependencies of service
type Settings struct {
	ServiceFactory   servicemgr.ServiceFactory
	Database         Database
	DirectoryManager DirectoryManager
	DownloadsManager DownloadsManager
	Remote           config.Remote
	Device           string
}

func (l LibraryService) GetTvSeriesUpdates(ctx context.Context, empty *emptypb.Empty, response *rms_library.GetTvSeriesUpdatesResponse) error {
	//TODO implement me
	panic("implement me")
}

func NewService(settings Settings) *LibraryService {
	// создаем клиента к rms-media-discovery
	tr := httptransport.New(settings.Remote.Host, settings.Remote.Path, []string{settings.Remote.Scheme})
	auth := httptransport.APIKeyAuth("X-Token", "header", settings.Device)
	discoveryClient := client.New(tr, strfmt.Default)

	return &LibraryService{
		f:                settings.ServiceFactory,
		auth:             auth,
		db:               settings.Database,
		cli:              discoveryClient,
		dir:              settings.DirectoryManager,
		dm:               settings.DownloadsManager,
		torrentToMovieID: map[string]string{},
	}
}
