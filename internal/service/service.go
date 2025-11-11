package service

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/dariamoshkina/shortify/internal/model"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrInvalidURL = errors.New("invalid URL")
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type URLRepository interface {
	GetByID(id string) (*model.URL, error)
	Store(url *model.URL) error
}

type ShortenerService struct {
	repo    URLRepository
	baseURL string
	length  int
}

func NewShortenerService(repo URLRepository, baseURL string, length int) *ShortenerService {
	return &ShortenerService{repo: repo, baseURL: baseURL, length: length}
}

func (s *ShortenerService) Shorten(original string) (string, error) {
	if original == "" {
		return "", ErrInvalidURL
	}
	if _, err := url.ParseRequestURI(original); err != nil {
		return "", ErrInvalidURL
	}

	var (
		id  string
		err error
	)
	for {
		id, err = randString(s.length)
		if err != nil {
			return "", fmt.Errorf("can't shorten URL: %w", err)
		}

		_, err = s.repo.GetByID(id)
		if errors.Is(err, ErrNotFound) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to check ID existence: %w", err)
		}
	}
	shortened := fmt.Sprintf("%s/%s", s.baseURL, id)

	if err = s.repo.Store(&model.URL{ID: id, Original: original, Shortened: shortened}); err != nil {
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

func randString(n int) (string, error) {
	result := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		result[i] = letters[num.Int64()]
	}
	return string(result), nil
}
