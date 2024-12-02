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

	// WaitTorrentReady means add data to directory only when torrent is downloaded
	WaitTorrentReady bool `json:"wait-torrent-ready"`
}

type Directories struct {
	// Downloads mean path to torrents content
	Downloads string

	// Layout means how dowloads organized. Common values:
	// "" (empty) - torrents stored as they downloaded
	// "%ID" - torrents organized by ID subdirectories
	// "data" - torrents organized in subfolder 'data'
	Layout string

	// Content means path to organized media
	Content string

	// Save original layout for internal torrent files
	// if false - the library decorate files
	SaveOriginalLayout bool `json:"save-original-layout"`
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
