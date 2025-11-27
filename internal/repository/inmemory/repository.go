package inmemory

import (
	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
)

type inMemoryRepository struct {
	storage map[string]model.URL
}

func NewInMemoryRepository() service.URLRepository {
	return &inMemoryRepository{
		storage: make(map[string]model.URL),
	}
}

func (r *inMemoryRepository) GetByID(id string) (*model.URL, error) {
	url, ok := r.storage[id]
	if !ok {
		return nil, service.ErrNotFound
	}
	return &url, nil
}

func (r *inMemoryRepository) Store(url *model.URL) error {
	r.storage[url.ID] = *url
	return nil
}
