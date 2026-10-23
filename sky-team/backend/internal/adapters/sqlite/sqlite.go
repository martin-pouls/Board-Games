// Package sqlite is the driven adapter that persists game state in an embedded SQLite
// database. The connection, migration runner and snapshot store are shared (shared/backend
// sqlitedb); this package owns Sky Team's migrations and the game <-> JSON snapshot mapping.
package sqlite

import (
	"context"
	"embed"
	"io/fs"

	"github.com/martin-pouls/shared/backend/sqlitedb"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// DB is the opened database.
type DB = sqlitedb.DB

// Open opens (creating if needed) the database file and applies pending migrations.
func Open(ctx context.Context, path string) (*DB, error) {
	migrations, err := fs.Sub(migrationsFS, "migrations")
	if err != nil {
		return nil, err
	}
	return sqlitedb.Open(ctx, path, migrations)
}
