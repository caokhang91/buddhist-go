package object

import (
	"errors"
	"fmt"
	"io"
	"os"
)

// NewBlobFromFile reads a file into a pooled blob.
func NewBlobFromFile(path string) (*Blob, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() < 0 {
		return nil, fmt.Errorf("invalid file size: %d", info.Size())
	}
	maxInt := int64(^uint(0) >> 1)
	if info.Size() > maxInt {
		return nil, fmt.Errorf("file too large to read: %d", info.Size())
	}
	size := int(info.Size())
	blob := NewBlob(size)
	if size == 0 {
		return blob, nil
	}
	if _, err := io.ReadFull(file, blob.Data); err != nil {
		_ = blob.Release()
		return nil, err
	}
	return blob, nil
}

// WriteToFile writes the blob contents to a file.
func (b *Blob) WriteToFile(path string) error {
	if b == nil {
		return errors.New("blob is nil")
	}
	return os.WriteFile(path, b.Data, 0644)
}
