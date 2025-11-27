package postgres

import (
	"context"
	"errors"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/jackc/pgx/v5"
)

type postgresRepository struct {
	conn *pgx.Conn
}

func NewPostgresRepository(connString string) service.URLRepository {
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
