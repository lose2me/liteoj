package db

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/liteoj/liteoj/backend/internal/config"
)

// Open connects to the configured database. Only sqlite and postgres are supported.
// For sqlite, the parent directory of the DSN is created if missing.
func Open(c *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	switch c.DBDriver {
	case "sqlite":
		if dir := filepath.Dir(c.DBDSN); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("mkdir %s: %w", dir, err)
			}
		}
		return gorm.Open(sqlite.Open(c.DBDSN), gormCfg)
	case "postgres":
		return gorm.Open(postgres.Open(c.DBDSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q", c.DBDriver)
	}
}
