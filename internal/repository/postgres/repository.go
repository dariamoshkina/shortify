package postgres

import (
	"context"
	"errors"
	"fmt"
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

func NewPostgresRepository(connString string) (service.URLRepository, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working dir: %w", err)
	}
	migrationsPath := "file://" + filepath.Join(wd, "migrations")

	m, err := migrate.New(migrationsPath, connString)
	if err != nil {
		return nil, fmt.Errorf("failed to create migrator: %w", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}
	if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
		return nil, fmt.Errorf("failed to close migrator: %w", err)
	}

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to db: %w", err)
	}

	return &postgresRepository{conn: conn}, nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*model.URL, error) {
	row := r.conn.QueryRow(ctx, "SELECT original, short FROM urls WHERE short_path = $1", id)

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

func (r *postgresRepository) Store(ctx context.Context, url model.URL) (*model.URL, error) {
	var res model.URL

	err := r.conn.QueryRow(ctx,
		`INSERT INTO urls (original, short, short_path)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (original) DO NOTHING
		 RETURNING original, short, short_path`,
		url.Original, url.Shortened, url.ID,
	).Scan(&res.Original, &res.Shortened, &res.ID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = r.conn.QueryRow(ctx,
				`SELECT original, short, short_path
				 FROM urls
				 WHERE original = $1`,
				url.Original,
			).Scan(&res.Original, &res.Shortened, &res.ID)
			if err != nil {
				return nil, err
			}
			return &res, nil
		}
		return nil, err
	}

	return nil, nil
}

func (r *postgresRepository) StoreMany(ctx context.Context, urls []model.URL) error {
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
		if _, err = tx.Exec(ctx, "batch_store", url.Original, url.Shortened, url.ID); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
