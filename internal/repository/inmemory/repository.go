package inmemory

import (
	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
)

type inMemoryURLRepository struct {
	storage map[string]model.URL
}

func NewInMemoryURLRepository() service.URLRepository {
	return &inMemoryURLRepository{
		storage: make(map[string]model.URL),
	}
}

func (r *inMemoryURLRepository) GetByID(id string) (*model.URL, error) {
	url, ok := r.storage[id]
	if !ok {
		return nil, service.ErrNotFound
	}
	return &url, nil
}

func (r *inMemoryURLRepository) Store(url *model.URL) error {
	r.storage[url.ID] = *url
	return nil
}
