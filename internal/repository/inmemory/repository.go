package inmemory

import (
	"context"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
	"github.com/google/uuid"
)

type inMemoryRepository struct {
	storage map[string]model.URL
}

func NewInMemoryRepository() service.URLRepository {
	return &inMemoryRepository{
		storage: make(map[string]model.URL),
	}
}

func (r *inMemoryRepository) GetByID(ctx context.Context, id string) (*model.URL, error) {
	url, ok := r.storage[id]
	if !ok {
		return nil, service.ErrNotFound
	}
	return &url, nil
}

func (r *inMemoryRepository) Store(ctx context.Context, url model.URL) (*model.URL, error) {
	for _, storedURL := range r.storage {
		if storedURL.Original == url.Original {
			return &storedURL, nil
		}
	}

	r.storage[url.ID] = url
	return nil, nil
}

func (r *inMemoryRepository) StoreMany(ctx context.Context, urls []model.URL) error {
	for _, url := range urls {
		r.storage[url.ID] = url
	}
	return nil
}

func (r *inMemoryRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]model.URL, error) {
	var result []model.URL
	for _, url := range r.storage {
		if url.UserID != nil && *url.UserID == userID {
			result = append(result, url)
		}
	}
	return result, nil
}
