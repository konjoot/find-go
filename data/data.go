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

// NewSource constructs concrete Source for srcType
func NewSource(srcType string) Source {
	return &source{srcType}
}

// Source is an interface for source reader
type Source interface {
	GetReadCloser(string) (io.ReadCloser, error)
}

// source is a URL\File Source implementation
type source struct {
	srcType string
}

// GetReadCloser returns io.ReadCloser from the target's source.
// Remember, you should allways call Close method at the end
func (s *source) GetReadCloser(target string) (io.ReadCloser, error) {
	switch s.srcType {
	case URL:
		resp, err := http.Get(target)
		if err != nil {
			return nil, err
		}
		return resp.Body, nil
	case File:
		file, err := os.Open(target)
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	return nil, errors.New("Unsupported data source type")
}
