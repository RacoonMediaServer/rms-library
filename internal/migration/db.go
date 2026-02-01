package migration

import (
	"context"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/db"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	rms_torrent "github.com/RacoonMediaServer/rms-packages/pkg/service/rms-torrent"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
)

type migratorFn func(f servicemgr.ServiceFactory) error

func (m *Migrator) migrateDatabaseV0ToV1(f servicemgr.ServiceFactory) error {
	torrCli := f.NewTorrent(false)

	movies, err := m.Database.SearchMovies(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("get all movies failed: %w", err)
	}

	for _, mov := range movies {
		if err = updateMovieLocations(m.Database, torrCli, mov); err != nil {
			return err
		}
	}

	return nil
}

func updateMovieLocations(d *db.Database, cli rms_torrent.RmsTorrentService, mov *model.Movie) error {
	updateDb := false
	for i := range mov.Torrents {
		t := &mov.Torrents[i]
		if t.Online || t.Location != "" {
			continue
		}
		info, err := cli.GetTorrentInfo(context.Background(), &rms_torrent.GetTorrentInfoRequest{Id: t.ID})
		if err != nil {
			return fmt.Errorf("get info about torrent '%s' of '%s' failed: %w", t.ID, mov.Info.Title, err)
		}
		t.Location = info.Location
		updateDb = true
	}

	if updateDb {
		return d.UpdateMovieContent(context.Background(), mov)
	}

	return nil
}
