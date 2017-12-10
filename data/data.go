package data

import (
	"errors"
	"io"
	"net/http"
	"os"
)

// Supported source types
const (
	URL  = "url"
	File = "file"
)

// NewSource constructs Source for srcType
func NewSource(srcType string) Source {
	return &source{srcType}
}

// Source is an interface for source reader
type Source interface {
	GetReader(string) (io.ReadCloser, error)
}

type source struct {
	srcType string
}

// GetReader returns io.ReadCloser from the target's source.
// Remember, you should call Close method at the end
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
