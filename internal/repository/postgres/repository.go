package postgres

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/jackc/pgx/v5"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/database/postgres"
	_ "github.com/golang-migrate/migrate/source/file"
)

type postgresRepository struct {
	conn *pgx.Conn
}

func NewPostgresRepository(connString string) service.URLRepository {
	wd, _ := os.Getwd()
	migrationsPath := "file://" + filepath.Join(wd, "migrations")

	m, err := migrate.New(
		migrationsPath,
		connString,
	)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to apply migrations: %v", err)
	}

	conn, _ := pgx.Connect(context.Background(), connString)

	return &postgresRepository{conn: conn}
}

func (r *postgresRepository) GetByID(id string) (*model.URL, error) {
	row := r.conn.QueryRow(context.Background(), "SELECT original, short FROM urls WHERE short_path = $1", id)

	var url model.URL
	err := row.Scan(&url.Original, &url.Shortened)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, service.ErrNotFound
		}
		return nil, err
	}

	return &url, nil
}

func (r *postgresRepository) Store(url *model.URL) error {
	_, err := r.conn.Exec(
		context.Background(),
		"INSERT INTO urls (original, short, short_path) VALUES ($1, $2, $3)", url.Original, url.Shortened, url.ID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (r *postgresRepository) StoreMany(urls []*model.URL) error {
	ctx := context.Background()
	tx, err := r.conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Prepare(ctx, "batch_store", "INSERT INTO urls (original, short, short_path) VALUES($1, $2, $3);")
	if err != nil {
		return err
	}

	for _, url := range urls {
		if _, err = r.conn.Exec(ctx, "batch_store", url.Original, url.Shortened, url.ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
