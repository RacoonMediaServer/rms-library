package migration

import (
	"context"
	"fmt"

	"github.com/RacoonMediaServer/rms-library/internal/config"
	"github.com/RacoonMediaServer/rms-library/internal/db"
	"github.com/RacoonMediaServer/rms-library/internal/model"
	"github.com/RacoonMediaServer/rms-packages/pkg/service/servicemgr"
	"go-micro.dev/v4/logger"
)

type Migrator struct {
	CurrentVersion string
	Database       *db.Database
	Config         config.Configuration

	mi *model.MetaInfo
}

func (m *Migrator) Run(f servicemgr.ServiceFactory) error {
	var err error

	m.mi, err = m.Database.GetMetaInfo(context.Background())
	if err != nil {
		return fmt.Errorf("get metainformation failed: %w", err)
	}

	if db.Version != m.mi.DatabaseVersion {
		logger.Warnf("Database schema version changed, migrate")
		if m.mi.DatabaseVersion > db.Version {
			return fmt.Errorf("cannot migrate database from future version: %d", m.mi.DatabaseVersion)
		}

		if err = m.migrateDatabase(f); err != nil {
			return fmt.Errorf("migrate database failed: %w", err)
		}
	}

	if m.CurrentVersion != m.mi.Version {
		m.mi.Version = m.CurrentVersion
		if err = m.Database.SetMetaInfo(context.Background(), *m.mi); err != nil {
			return fmt.Errorf("update meta information failed: %w", err)
		}
	}

	return nil
}

func (m *Migrator) migrateDatabase(f servicemgr.ServiceFactory) error {
	migrations := m.getMigrations()
	for cur := m.mi.DatabaseVersion; cur < db.Version; cur++ {
		if err := migrations[cur](f); err != nil {
			return fmt.Errorf("from %d to %d: %w", cur, cur+1, err)
		}
		m.mi.DatabaseVersion = cur + 1
		if err := m.Database.SetMetaInfo(context.Background(), *m.mi); err != nil {
			return fmt.Errorf("update meta information failed: %w", err)
		}
	}
	return nil
}

func (m *Migrator) getMigrations() []migratorFn {
	return []migratorFn{
		m.migrateDatabaseV0ToV1,
		m.migrateDatabaseV1ToV2,
	}
}
