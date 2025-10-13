package service

import (
	"fmt"
	"math/rand"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/repository/interfaces"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type ShortenerService struct {
	repo    interfaces.URLRepository
	baseURL string
}

func NewShortenerService(repo interfaces.URLRepository, baseURL string) *ShortenerService {
	return &ShortenerService{repo: repo, baseURL: baseURL}
}

func (s *ShortenerService) Shorten(urlString string) (string, error) {
	id := randString(6)
	shortened := fmt.Sprintf("%s/%s", s.baseURL, id)

	err := s.repo.Store(&model.URL{
		ID:        id,
		Original:  urlString,
		Shortened: shortened,
	})
	if err != nil {
		return "", err
	}

	return shortened, nil
}

func (s *ShortenerService) Restore(id string) (string, error) {
	url, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}
	return url.Original, nil
}

func randString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
