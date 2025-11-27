package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type DBHandler struct {
	dbConnString string
}

func NewDBHandler(dbConnString string) *DBHandler {
	return &DBHandler{dbConnString: dbConnString}
}

func (h *DBHandler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.PingHandler)

	return r
}

func (h *DBHandler) PingHandler(res http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	database, err := pgx.Connect(ctx, h.dbConnString)

	if err != nil {
		http.Error(res, "failed to connect to database", http.StatusInternalServerError)
		return
	}
	defer database.Close(ctx)

	res.WriteHeader(http.StatusOK)
}
