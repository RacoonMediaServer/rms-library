package movies

import (
	"context"
	"math/rand"
	"time"

	"github.com/RacoonMediaServer/rms-library/internal/config"
	"github.com/RacoonMediaServer/rms-library/internal/lock"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-library/internal/schedule"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

const searchTorrentsLimit uint = 10

// MoviesService is a service API handler
type MoviesService struct {
	f     servicemgr.ServiceFactory
	auth  runtime.ClientAuthInfoWriter
	db    Database
	cli   *client.Client
	dir   DirectoryManager
	dm    DownloadsManager
	sched Scheduler
	lk    lock.Locker
	pub   micro.Event
}

// Get implements rms_library.MoviesHandler.
func (l *MoviesService) Get(ctx context.Context, req *rms_library.MoviesGetRequest, resp *rms_library.MoviesGetResponse) error {
	mov, err := l.db.GetMovie(ctx, model.ID(req.Id))
	if err != nil {
		return err
	}

	resp.Info = &mov.Info
	return nil
}

// Settings holds all dependencies of service
type Settings struct {
	ServiceFactory   servicemgr.ServiceFactory
	Database         Database
	DirectoryManager DirectoryManager
	DownloadsManager DownloadsManager
	Remote           config.Remote
	Device           string
	Scheduler        Scheduler
	Locker           lock.Locker
	Publisher        micro.Event
}

func NewService(settings Settings) *MoviesService {
	// создаем клиента к rms-media-discovery
	tr := httptransport.New(settings.Remote.Host, settings.Remote.Path, []string{settings.Remote.Scheme})
	auth := httptransport.APIKeyAuth("X-Token", "header", settings.Device)
	discoveryClient := client.New(tr, strfmt.Default)

	l := &MoviesService{
		f:     settings.ServiceFactory,
		auth:  auth,
		db:    settings.Database,
		cli:   discoveryClient,
		dir:   settings.DirectoryManager,
		dm:    settings.DownloadsManager,
		sched: settings.Scheduler,
		lk:    settings.Locker,
		pub:   settings.Publisher,
	}

	return l
}

func (l MoviesService) Initialize() error {
	movies, err := l.db.SearchMovies(context.Background(), nil)
	if err != nil {
		return err
	}

	for _, mov := range movies {
		logger.Debugf("Movie found: %s", mov.Title)

		l.dir.CreateMovieLayout(mov)

		// start watcher
		task := schedule.Task{
			Group: mov.ID.String(),
			Fn: schedule.GetPeriodicWrapper(
				logger.Fields(map[string]interface{}{
					"op":    "movieWatcher",
					"id":    mov.ID.String(),
					"title": mov.Info.Title,
				}),
				watchInterval,
				func(log logger.Logger, ctx context.Context) error {
					return l.asyncWatch(log, ctx, mov.ID)
				},
			),
		}
		task.After(time.Duration(rand.Intn(240)) * time.Second)
		l.sched.Add(&task)
	}

	return nil
}
