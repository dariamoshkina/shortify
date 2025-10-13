package interfaces

import "github.com/dariamoshkina/shortify/internal/model"

type URLRepository interface {
	GetByID(id string) (*model.URL, error)
	Store(url *model.URL) error
}
