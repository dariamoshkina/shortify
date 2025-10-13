package interfaces

import "github.com/dariamoshkina/shortify/internal/model"

type UrlRepository interface {
	GetByID(id string) (*model.Url, error)
	Store(url *model.Url) error
}
