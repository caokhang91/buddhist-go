//go:build !windows

package object

import (
	"fmt"
	"os"
	"syscall"
)

// MmapFile maps a file into memory and returns a Blob referencing it.
func MmapFile(path string, writable bool) (*Blob, error) {
	flag := os.O_RDONLY
	prot := syscall.PROT_READ
	if writable {
		flag = os.O_RDWR
		prot |= syscall.PROT_WRITE
	}

	file, err := os.OpenFile(path, flag, 0)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() == 0 {
		return NewBlob(0), nil
	}
	if info.Size() < 0 {
		return nil, fmt.Errorf("invalid file size: %d", info.Size())
	}
	maxInt := int64(^uint(0) >> 1)
	if info.Size() > maxInt {
		return nil, fmt.Errorf("file too large to mmap: %d", info.Size())
	}

	data, err := syscall.Mmap(int(file.Fd()), 0, int(info.Size()), prot, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}

	return &Blob{
		Data:     data,
		mmapped:  true,
		ownsData: true,
	}, nil
}

// Unmap releases a memory-mapped blob.
func (b *Blob) Unmap() error {
	if b == nil {
		return fmt.Errorf("blob is nil")
	}
	if !b.mmapped {
		return fmt.Errorf("blob is not memory-mapped")
	}
	if !b.ownsData {
		return fmt.Errorf("blob does not own memory map")
	}
	if err := syscall.Munmap(b.Data); err != nil {
		return err
	}
	b.Data = nil
	b.mmapped = false
	b.ownsData = false
	return nil
}
