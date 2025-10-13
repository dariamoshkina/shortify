package inmemory

import (
	"errors"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/repository/interfaces"
)

type inMemoryURLRepository struct {
	storage map[string]model.URL
}

func NewInMemoryURLRepository() interfaces.URLRepository {
	return &inMemoryURLRepository{
		storage: make(map[string]model.URL),
	}
}

func (r *inMemoryURLRepository) GetByID(id string) (*model.URL, error) {
	url, ok := r.storage[id]
	if !ok {
		return nil, errors.New("url not found")
	}
	return &url, nil
}

func (r *inMemoryURLRepository) Store(url *model.URL) error {
	r.storage[url.ID] = *url
	return nil
}
