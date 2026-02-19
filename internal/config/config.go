package config

import "github.com/RacoonMediaServer/rms-packages/pkg/configuration"

// Remote is settings for connection to rms-bot-server service
type Remote struct {
	Scheme string
	Host   string
	Port   int
	Path   string
}

// Configuration represents entire service configuration
type Configuration struct {
	// MongoDB connection string
	Database string

	// Device API key
	Device string

	// Directories are paths to media content
	Directories Directories

	// Remote is settings to connect to the Remote Server
	Remote Remote
}

type Directories struct {
	// Downloads mean path to torrents content
	Downloads string

	// Content means path to organized media
	Content string

	// Path to directory for store watch-list torrent files
	WatchList string `json:"watch-list"`
}

var config Configuration

// Load open and parses configuration file
func Load(configFilePath string) error {
	return configuration.Load(configFilePath, &config)
}

// Config returns loaded configuration
func Config() Configuration {
	return config
}
