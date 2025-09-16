package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Up aplica todas las migraciones pendientes.
func Up(ctx context.Context, dsn string) error {
	if dsn == "" {
		return fmt.Errorf("empty DSN")
	}
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	return goose.UpContext(ctx, db, "migrations")
}
