package lowl

import (
	"bytes"
	"io"
	"os"
)

type Reader struct {
	rdr io.Reader
}

func NewReader(name string) (*Reader, error) {
	data, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &Reader{rdr: bytes.NewReader(data)}, nil
}
