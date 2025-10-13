package in_memory

import (
	"errors"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/repository/interfaces"
)

type inMemoryUrlRepository struct {
	storage map[string]model.Url
}

func NewInMemoryUrlRepository() interfaces.UrlRepository {
	return &inMemoryUrlRepository{
		storage: make(map[string]model.Url),
	}
}

func (r *inMemoryUrlRepository) GetByID(id string) (*model.Url, error) {
	url, ok := r.storage[id]
	if !ok {
		return nil, errors.New("url not found")
	}
	return &url, nil
}

func (r *inMemoryUrlRepository) Store(url *model.Url) error {
	r.storage[url.ID] = *url
	return nil
}
