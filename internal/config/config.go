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

	// Directory is a base media directory
	Directory string

	// Remote is settings to connect to the Remote Server
	Remote Remote

	// FixTorrentPath means use 'data' instead of torrent ID in the path of content
	FixTorrentPath bool `json:"fix-torrent-path"`
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
