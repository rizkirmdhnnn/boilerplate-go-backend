package migrator

import (
	"errors"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/rs/zerolog/log"
)

// Run executes all pending migrations using golang-migrate.
// dsn must be a full postgres:// connection string.
func Run(dsn, dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return fmt.Errorf("migrations directory %q not found", dir)
	}

	sourceURL := fmt.Sprintf("file://%s", dir)
	m, err := migrate.New(sourceURL, dsn)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("no new migrations to apply")
			return nil
		}
		return fmt.Errorf("run migrations: %w", err)
	}

	log.Info().Msg("all migrations applied successfully")
	return nil
}
