package migrator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"boilerplate/pkg/database"

	"github.com/rs/zerolog/log"
)

// Run reads all .sql files from the given directory sorted by name
// and executes them against the database pool.
func Run(ctx context.Context, db *database.Pool, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations directory %q: %w", dir, err)
	}

	// Filter and sort .sql files
	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, fname := range files {
		path := filepath.Join(dir, fname)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %q: %w", fname, err)
		}

		sql := strings.TrimSpace(string(content))
		if sql == "" {
			continue
		}

		if _, err := db.Exec(ctx, sql); err != nil {
			return fmt.Errorf("execute migration %q: %w", fname, err)
		}

		log.Info().Str("migration", fname).Msg("applied")
	}

	return nil
}
