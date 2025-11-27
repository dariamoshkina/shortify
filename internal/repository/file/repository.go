package file

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"github.com/dariamoshkina/shortify/internal/model"
	"github.com/dariamoshkina/shortify/internal/service"
)

var (
	errOpenFile  = errors.New("can't open file")
	errReadFile  = errors.New("can't read file")
	errWriteFile = errors.New("can't write file")
)

type fileRepository struct {
	filename string
}

func NewFileRepository(filename string) service.URLRepository {
	return &fileRepository{
		filename: filename,
	}
}

func (f fileRepository) GetByID(id string) (*model.URL, error) {
	file, err := os.OpenFile(f.filename, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, errOpenFile
	}
	defer file.Close()

	var urls []model.URL
	dec := json.NewDecoder(file)
	if err = dec.Decode(&urls); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, service.ErrNotFound
		}
		return nil, errReadFile
	}

	for _, url := range urls {
		if url.ID == id {
			return &url, nil
		}
	}
	return nil, service.ErrNotFound
}

func (f fileRepository) Store(url *model.URL) error {
	file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errOpenFile
	}
	defer file.Close()

	var urls []model.URL
	dec := json.NewDecoder(file)
	if err = dec.Decode(&urls); err != nil {
		if !errors.Is(err, io.EOF) {
			return errReadFile
		}
	}
	if _, err = file.Seek(0, 0); err != nil {
		return errWriteFile
	}

	urls = append(urls, *url)
	enc := json.NewEncoder(file)
	if err = enc.Encode(urls); err != nil {
		return errWriteFile
	}

	return nil
}

func (f fileRepository) StoreMany(urls []*model.URL) error {
	file, err := os.OpenFile(f.filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return errOpenFile
	}
	defer file.Close()

	var existingURLs []model.URL
	dec := json.NewDecoder(file)
	if err = dec.Decode(&existingURLs); err != nil {
		if !errors.Is(err, io.EOF) {
			return errReadFile
		}
	}
	if _, err = file.Seek(0, 0); err != nil {
		return errWriteFile
	}

	for _, url := range urls {
		existingURLs = append(existingURLs, model.URL{ID: url.ID, Original: url.Original, Shortened: url.Shortened})
	}
	enc := json.NewEncoder(file)
	if err = enc.Encode(urls); err != nil {
		return errWriteFile
	}

	return nil
}
