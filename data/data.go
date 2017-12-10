package data

import (
	"errors"
	"io"
	"net/http"
	"os"
)

const (
	URL  = "url"
	File = "file"
)

func NewSource(srcType string) Source {
	return &source{srcType}
}

type Source interface {
	GetReader(string) (io.ReadCloser, error)
}

type source struct {
	srcType string
}

func (s *source) GetReader(target string) (io.ReadCloser, error) {
	switch s.srcType {
	case URL:
		resp, err := http.Get(target)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	case File:
		return os.Open(target)
	}
	return nil, errors.New("Unsupported data source type")
}
