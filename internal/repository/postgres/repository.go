package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) service.URLRepository {
	return &postgresRepository{pool: pool}
}

func RunMigrations(connString string) error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working dir: %w", err)
	}
	migrationsPath := "file://" + filepath.Join(wd, "migrations")

	m, err := migrate.New(migrationsPath, connString)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	if err = m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	if srcErr, dbErr := m.Close(); srcErr != nil || dbErr != nil {
		return fmt.Errorf("failed to close migrator: %w", err)
	}

	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*model.URL, error) {
	row := r.pool.QueryRow(ctx, "SELECT original, short FROM urls WHERE short_path = $1", id)

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

	err := r.pool.QueryRow(ctx,
		`INSERT INTO urls (original, short, short_path, user_id)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (original) DO NOTHING
		 RETURNING original, short, short_path, user_id`,
		url.Original, url.Shortened, url.ID, url.UserID,
	).Scan(&res.Original, &res.Shortened, &res.ID, &res.UserID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = r.pool.QueryRow(ctx,
				`SELECT original, short, short_path, user_id
				 FROM urls
				 WHERE original = $1`,
				url.Original,
			).Scan(&res.Original, &res.Shortened, &res.ID, &res.UserID)
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
	batch := &pgx.Batch{}
	for _, u := range urls {
		batch.Queue(
			`INSERT INTO urls (original, short, short_path, user_id)
             VALUES ($1, $2, $3, $4);`,
			u.Original, u.Shortened, u.ID, u.UserID,
		)
	}

	br := r.pool.SendBatch(ctx, batch)
	defer br.Close()

	for range urls {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	return nil
}

func (r *postgresRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.URL, error) {
	var result []model.URL

	rows, err := r.pool.Query(ctx, "SELECT original, short FROM urls WHERE user_id = $1", userID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url model.URL
		if err = rows.Scan(&url.Original, &url.Shortened); err != nil {
			return nil, err
		}
		result = append(result, url)
	}

	return result, nil
}
