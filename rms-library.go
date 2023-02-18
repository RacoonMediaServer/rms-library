package main

import (
	"fmt"
	"github.com/RacoonMediaServer/rms-library/internal/config"
	"github.com/RacoonMediaServer/rms-library/internal/db"
	libraryService "github.com/RacoonMediaServer/rms-library/internal/service"
	"github.com/RacoonMediaServer/rms-media-discovery/pkg/client/client"
	rms_library "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-library"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	"github.com/urfave/cli/v2"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
)

var Version = "v0.0.0"

const serviceName = "rms-library"
const discoveryEndpoint = "136.244.108.126"

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

	f := servicemgr.NewServiceFactory(service)

	database, err := db.Connect(cfg.Database)
	if err != nil {
		logger.Fatalf("Connect to database failed: %s", err)
	}
	logger.Info("Connected to database")

	// создаем клиента к Remote-сервису rms-media-discovery
	tr := httptransport.New(discoveryEndpoint, "/media", client.DefaultSchemes)
	auth := httptransport.APIKeyAuth("X-Token", "header", cfg.Device)
	discoveryClient := client.New(tr, strfmt.Default)

	lib := libraryService.NewService(database, f, discoveryClient, auth)

	//регистрируем хендлеры
	if err := rms_library.RegisterRmsLibraryHandler(service.Server(), lib); err != nil {
		logger.Fatalf("Register service failed: %s", err)
	}

	if err := service.Run(); err != nil {
		logger.Fatalf("Run service failed: %s", err)
	}
}
