package main

import (
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/config"
	"github.com/RacoonMediaServer/rms-library/internal/db"
	"github.com/RacoonMediaServer/rms-library/internal/downloads"
	"github.com/RacoonMediaServer/rms-library/internal/service/movies"
	"github.com/RacoonMediaServer/rms-library/internal/storage"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"

	// Plugins
	_ "github.com/go-micro/plugins/v4/registry/etcd"
)

var Version = "v0.0.0"

const serviceName = "rms-library"

func main() {
	logger.Infof("%s %s", serviceName, Version)
	defer logger.Info("DONE.")

	useDebug := false

	service := micro.NewService(
		micro.Name(serviceName),
		micro.Version(Version),
		micro.Flags(
			&cli.BoolFlag{
				Name:        "verbose",
				Aliases:     []string{"debug"},
				Usage:       "debug log level",
				Value:       false,
				Destination: &useDebug,
			},
		),
	)

	service.Init(
		micro.Action(func(context *cli.Context) error {
			configFile := fmt.Sprintf("/etc/rms/%s.json", serviceName)
			if context.IsSet("config") {
				configFile = context.String("config")
			}
			return config.Load(configFile)
		}),
	)

	if useDebug {
		_ = logger.Init(logger.WithLevel(logger.DebugLevel))
	}

	cfg := config.Config()

	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.Fatalf("Connect to database failed: %s", err)
	}
	logger.Info("Connected to database")

	// фабрика коннекторов к другим сервисам
	f := servicemgr.NewServiceFactory(service)

	// создаем структуру директорий
	dirManager, err := storage.NewManager(cfg.Directories)
	if err != nil {
		logger.Fatalf("Cannot initialize directory manager: %s", err)
	}

	// создаем менеджер закачек
	downloadManager := downloads.NewManager(f.NewTorrent(), database, dirManager, cfg.WaitTorrentReady)
	if err = downloadManager.Initialize(); err != nil {
		logger.Fatalf("Cannot initialize downloads manager: %s", err)
	}

	settings := movies.Settings{
		ServiceFactory:   f,
		Database:         database,
		DirectoryManager: dirManager,
		DownloadsManager: downloadManager,
		Remote:           cfg.Remote,
		Device:           cfg.Device,
	}

	moviesService := movies.NewService(settings)

	// подписываемся на события от торрентов
	if err = moviesService.Subscribe(service.Server()); err != nil {
		logger.Fatalf("Subscribe failed: %s", err)
	}

	//регистрируем хендлеры
	if err = rms_library.RegisterMoviesHandler(service.Server(), moviesService); err != nil {
		logger.Fatalf("Register service failed: %s", err)
	}

	if err = service.Run(); err != nil {
		logger.Fatalf("Run service failed: %s", err)
	}
}
