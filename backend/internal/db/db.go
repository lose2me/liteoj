package db

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/liteoj/liteoj/backend/internal/config"
)

// Open connects to the configured database. Only sqlite and postgres are supported.
// For sqlite, the parent directory of the DSN is created if missing and a set
// of defensive PRAGMAs is applied: WAL journal + NORMAL sync for better read/write
// concurrency under queue_workers=1; busy_timeout to let concurrent writers
// retry briefly rather than error out; foreign_keys for integrity; temp_store
// MEMORY to skip the temp file roundtrip for aggregation queries (ranking).
// PRAGMA failures are logged but non-fatal — they are pure optimizations.
func Open(c *config.Config) (*gorm.DB, error) {
	gormCfg := &gorm.Config{Logger: logger.Default.LogMode(logger.Warn)}
	switch c.DBDriver {
	case "sqlite":
		if dir := filepath.Dir(c.DBDSN); dir != "" && dir != "." {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("mkdir %s: %w", dir, err)
			}
		}
		gdb, err := gorm.Open(sqlite.Open(c.DBDSN), gormCfg)
		if err != nil {
			return nil, err
		}
		applySQLitePragmas(gdb)
		return gdb, nil
	case "postgres":
		return gorm.Open(postgres.Open(c.DBDSN), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported DB_DRIVER %q", c.DBDriver)
	}
}

// applySQLitePragmas runs a handful of defensive PRAGMAs right after Open.
// Order matters: journal_mode has to come before synchronous for the WAL
// pairing to stick. Each statement is logged so a misconfigured filesystem
// (e.g. WAL on a network mount) surfaces at startup rather than as weird
// concurrent-write failures later.
func applySQLitePragmas(gdb *gorm.DB) {
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
		"PRAGMA temp_store=MEMORY",
	}
	for _, p := range pragmas {
		if err := gdb.Exec(p).Error; err != nil {
			log.Printf("db: %s failed: %v", p, err)
		}
	}
}
