package service

import (
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/config"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/models"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

func init() {
	// this is a trick for increase timeouts for all clients to OpenAPI
	httptransport.DefaultTimeout = 120 * time.Second
}

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
	torrentToResult  map[string]*models.SearchTorrentsResult
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

func NewService(settings Settings) *LibraryService {
	// создаем клиента к rms-media-discovery
	tr := httptransport.New(settings.Remote.Host, settings.Remote.Path, []string{settings.Remote.Scheme})
	auth := httptransport.APIKeyAuth("X-Token", "header", settings.Device)
	discoveryClient := client.New(tr, strfmt.Default)

	l := &LibraryService{
		f:                settings.ServiceFactory,
		auth:             auth,
		db:               settings.Database,
		cli:              discoveryClient,
		dir:              settings.DirectoryManager,
		dm:               settings.DownloadsManager,
		torrentToMovieID: map[string]string{},
		torrentToResult:  map[string]*models.SearchTorrentsResult{},
	}

	go l.checkAvailableUpdates()

	return l
}
