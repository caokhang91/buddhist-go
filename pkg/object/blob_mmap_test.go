//go:build !windows

package object

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func TestBlobMmapRead(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mmap-read.bin")
	data := []byte{1, 2, 3, 4}

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	blob, err := MmapFile(path, false)
	if err != nil {
		t.Fatalf("MmapFile failed: %v", err)
	}
	if !bytes.Equal(blob.Data, data) {
		t.Fatalf("unexpected mmap data: %v", blob.Data)
	}
	if err := blob.Unmap(); err != nil {
		t.Fatalf("Unmap failed: %v", err)
	}
	if err := blob.Unmap(); err == nil {
		t.Fatalf("expected error on double unmap")
	}
}

func TestBlobMmapWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "mmap-write.bin")
	data := make([]byte, 8)

	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	blob, err := MmapFile(path, true)
	if err != nil {
		t.Fatalf("MmapFile failed: %v", err)
	}
	if err := blob.WriteInt64(0, 0x1122334455667788); err != nil {
		t.Fatalf("WriteInt64 failed: %v", err)
	}
	if err := blob.Unmap(); err != nil {
		t.Fatalf("Unmap failed: %v", err)
	}

	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	value := int64(binary.LittleEndian.Uint64(onDisk))
	if value != 0x1122334455667788 {
		t.Fatalf("unexpected value after mmap write: %x", value)
	}
}
