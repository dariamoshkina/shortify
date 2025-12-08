package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/samber/lo"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrInvalidURL = errors.New("invalid URL")
	ErrDuplicate  = errors.New("duplicate")
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type URLRepository interface {
	GetByID(ctx context.Context, id string) (*model.URL, error)
	Store(ctx context.Context, url model.URL) (*model.URL, error)
	StoreMany(ctx context.Context, urls []model.URL) error
}

type ShortenerService struct {
	repo    URLRepository
	baseURL string
	length  int
}

func NewShortenerService(repo URLRepository, baseURL string, length int) *ShortenerService {
	return &ShortenerService{repo: repo, baseURL: baseURL, length: length}
}

func (s *ShortenerService) Shorten(ctx context.Context, original string) (string, error) {
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

		_, err = s.repo.GetByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to check ID existence: %w", err)
		}
	}
	shortened := fmt.Sprintf("%s/%s", s.baseURL, id)

	existingURL, err := s.repo.Store(ctx, model.URL{ID: id, Original: original, Shortened: shortened})
	if err != nil {
		return "", err
	}
	if existingURL != nil {
		return existingURL.Shortened, ErrDuplicate
	}

	return shortened, nil
}

func (s *ShortenerService) ShortenMany(ctx context.Context, urls []model.BatchURL) ([]*model.BatchURL, error) {
	if len(urls) == 0 {
		return nil, ErrInvalidURL
	}
	if !lo.EveryBy(urls, func(u model.BatchURL) bool {
		_, err := url.ParseRequestURI(u.Original)
		return err == nil
	}) {
		return nil, ErrInvalidURL
	}

	var (
		id      string
		err     error
		result  = make([]*model.BatchURL, 0, len(urls))
		toStore = make([]model.URL, 0, len(urls))
	)
	for _, sourceURL := range urls {
		for {
			id, err = randString(s.length)
			if err != nil {
				return nil, fmt.Errorf("can't shorten URL: %w", err)
			}

			_, err = s.repo.GetByID(ctx, id)
			if errors.Is(err, ErrNotFound) {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to check ID existence: %w", err)
			}
		}
		shortened := fmt.Sprintf("%s/%s", s.baseURL, id)

		toStore = append(toStore, model.URL{ID: id, Original: sourceURL.Original, Shortened: shortened})
		result = append(result, &model.BatchURL{CorrID: sourceURL.CorrID, Shortened: shortened})
	}

	if err = s.repo.StoreMany(ctx, toStore); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *ShortenerService) Restore(ctx context.Context, id string) (string, error) {
	url, err := s.repo.GetByID(ctx, id)
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
