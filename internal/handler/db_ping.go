package handler

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type DBPinger struct {
	pool   *pgxpool.Pool
	logger *zap.SugaredLogger
}

func NewDBPinger(pool *pgxpool.Pool, logger *zap.SugaredLogger) *DBPinger {
	return &DBPinger{pool: pool, logger: logger}
}

func (h *DBPinger) PingHandler(res http.ResponseWriter, req *http.Request) {
	if err := h.pool.Ping(req.Context()); err != nil {
		h.logger.Error("failed to ping database", zap.Error(err))
		http.Error(res, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusOK)
}
