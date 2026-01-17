//go:build windows

package object

import "errors"

// MmapFile is not supported on Windows builds.
func MmapFile(path string, writable bool) (*Blob, error) {
	return nil, errors.New("memory mapping not supported on windows")
}

// Unmap releases a memory-mapped blob (unsupported on Windows).
func (b *Blob) Unmap() error {
	if b == nil {
		return errors.New("blob is nil")
	}
	if !b.mmapped {
		return errors.New("blob is not memory-mapped")
	}
	return errors.New("memory mapping not supported on windows")
}
