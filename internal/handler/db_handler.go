package handler

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type DBHandler struct {
	ctx          context.Context
	dbConnString string
}

func NewDBHandler(ctx context.Context, dbConnString string) *DBHandler {
	return &DBHandler{ctx: ctx, dbConnString: dbConnString}
}

func (h *DBHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.PingHandler)

	return r
}

func (h *DBHandler) PingHandler(res http.ResponseWriter, req *http.Request) {
	database, err := pgx.Connect(h.ctx, h.dbConnString)

	if err != nil {
		http.Error(res, "failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close(h.ctx)

	res.WriteHeader(http.StatusOK)
}
