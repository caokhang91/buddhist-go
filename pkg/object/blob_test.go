package object

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestBlobReadWriteIntFloat(t *testing.T) {
	blob := NewBlob(16)

	if err := blob.WriteInt64(0, 42); err != nil {
		t.Fatalf("WriteInt64 failed: %v", err)
	}
	if err := blob.WriteFloat64(8, 3.25); err != nil {
		t.Fatalf("WriteFloat64 failed: %v", err)
	}

	intVal, err := blob.ReadInt64(0)
	if err != nil {
		t.Fatalf("ReadInt64 failed: %v", err)
	}
	if intVal != 42 {
		t.Fatalf("unexpected int value: %d", intVal)
	}

	floatVal, err := blob.ReadFloat64(8)
	if err != nil {
		t.Fatalf("ReadFloat64 failed: %v", err)
	}
	if math.Abs(floatVal-3.25) > 1e-9 {
		t.Fatalf("unexpected float value: %f", floatVal)
	}

	if _, err := blob.ReadInt64(10); err == nil {
		t.Fatalf("expected out-of-range error")
	}
}

func TestBlobSliceZeroCopy(t *testing.T) {
	blob := NewBlobFromBytes([]byte{1, 2, 3, 4})
	sub, err := blob.SubBlob(1, 3)
	if err != nil {
		t.Fatalf("SubBlob failed: %v", err)
	}
	if !bytes.Equal(sub.Data, []byte{2, 3}) {
		t.Fatalf("unexpected slice data: %v", sub.Data)
	}

	blob.Data[1] = 9
	if sub.Data[0] != 9 {
		t.Fatalf("expected zero-copy slice to reflect changes")
	}

	if _, err := blob.SubBlob(3, 5); err == nil {
		t.Fatalf("expected out-of-range error")
	}
}

func TestBlobRelease(t *testing.T) {
	blob := NewBlob(4)
	if err := blob.Release(); err != nil {
		t.Fatalf("Release failed: %v", err)
	}
	if blob.Data != nil {
		t.Fatalf("expected blob data to be nil after release")
	}
	if err := blob.Release(); err != nil {
		t.Fatalf("Release should be idempotent: %v", err)
	}
}

func TestBlobFileIO(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "blob.bin")
	data := []byte{7, 8, 9}

	blob := NewBlobFromBytes(data)
	if err := blob.WriteToFile(path); err != nil {
		t.Fatalf("WriteToFile failed: %v", err)
	}

	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if !bytes.Equal(onDisk, data) {
		t.Fatalf("unexpected file data: %v", onDisk)
	}

	loaded, err := NewBlobFromFile(path)
	if err != nil {
		t.Fatalf("NewBlobFromFile failed: %v", err)
	}
	if !bytes.Equal(loaded.Data, data) {
		t.Fatalf("unexpected blob data: %v", loaded.Data)
	}
}

func TestBlobBuiltins(t *testing.T) {
	newBuiltin := GetBuiltinByName("blob_new")
	if newBuiltin == nil {
		t.Fatalf("blob_new builtin not found")
	}
	newResult := newBuiltin.Fn(&Integer{Value: 8})
	blob, ok := newResult.(*Blob)
	if !ok {
		t.Fatalf("expected Blob, got %T", newResult)
	}
	if len(blob.Data) != 8 {
		t.Fatalf("expected blob length 8, got %d", len(blob.Data))
	}

	writeBuiltin := GetBuiltinByName("blob_write_int")
	readBuiltin := GetBuiltinByName("blob_read_int")
	if writeBuiltin == nil || readBuiltin == nil {
		t.Fatalf("blob read/write builtins not found")
	}

	writeResult := writeBuiltin.Fn(blob, &Integer{Value: 0}, &Integer{Value: 123})
	if _, ok := writeResult.(*Error); ok {
		t.Fatalf("unexpected error from blob_write_int: %v", writeResult.Inspect())
	}

	readResult := readBuiltin.Fn(blob, &Integer{Value: 0})
	readInt, ok := readResult.(*Integer)
	if !ok || readInt.Value != 123 {
		t.Fatalf("unexpected read result: %v", readResult.Inspect())
	}

	errorResult := readBuiltin.Fn(blob, &Integer{Value: -1})
	if _, ok := errorResult.(*Error); !ok {
		t.Fatalf("expected error for negative offset")
	}

	sliceBuiltin := GetBuiltinByName("blob_slice")
	if sliceBuiltin == nil {
		t.Fatalf("blob_slice builtin not found")
	}
	sliceBlob := NewBlobFromBytes([]byte{1, 2, 3, 4})
	sliceResult := sliceBuiltin.Fn(sliceBlob, &Integer{Value: -3}, &Integer{Value: -1})
	sliced, ok := sliceResult.(*Blob)
	if !ok {
		t.Fatalf("expected Blob slice, got %T", sliceResult)
	}
	if !bytes.Equal(sliced.Data, []byte{2, 3}) {
		t.Fatalf("unexpected sliced data: %v", sliced.Data)
	}
}
