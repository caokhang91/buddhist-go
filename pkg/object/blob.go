package object

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"sync"
)

var blobPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0)
	},
}

// Blob represents a binary blob backed by a byte slice.
type Blob struct {
	Data     []byte
	pooled   bool
	mmapped  bool
	ownsData bool
}

func (b *Blob) Type() ObjectType { return BLOB_OBJ }
func (b *Blob) Inspect() string  { return fmt.Sprintf("blob(len=%d)", len(b.Data)) }

// NewBlob allocates a zeroed blob with the given size.
func NewBlob(size int) *Blob {
	if size < 0 {
		size = 0
	}
	buf := blobPool.Get().([]byte)
	if cap(buf) < size {
		buf = make([]byte, size)
	} else {
		buf = buf[:size]
		for i := range buf {
			buf[i] = 0
		}
	}
	return &Blob{Data: buf, pooled: true, ownsData: true}
}

// NewBlobFromBytes copies input data into a pooled blob.
func NewBlobFromBytes(data []byte) *Blob {
	if len(data) == 0 {
		return NewBlob(0)
	}
	blob := NewBlob(len(data))
	copy(blob.Data, data)
	return blob
}

// Release returns pooled memory back to the pool.
func (b *Blob) Release() error {
	if b == nil {
		return errors.New("blob is nil")
	}
	if b.mmapped {
		return errors.New("blob is memory-mapped; use blob_unmap")
	}
	if !b.pooled || !b.ownsData {
		return nil
	}
	blobPool.Put(b.Data[:0])
	b.Data = nil
	b.ownsData = false
	return nil
}

// SubBlob returns a zero-copy slice of the blob.
func (b *Blob) SubBlob(start, end int) (*Blob, error) {
	if b == nil {
		return nil, errors.New("blob is nil")
	}
	if start < 0 || end < 0 || start > end || end > len(b.Data) {
		return nil, fmt.Errorf("slice out of range: %d:%d", start, end)
	}
	return &Blob{
		Data:     b.Data[start:end],
		pooled:   b.pooled,
		mmapped:  b.mmapped,
		ownsData: false,
	}, nil
}

// ReadInt64 reads an int64 from the blob using little-endian.
func (b *Blob) ReadInt64(offset int) (int64, error) {
	if err := b.checkRange(offset, 8); err != nil {
		return 0, err
	}
	return int64(binary.LittleEndian.Uint64(b.Data[offset:])), nil
}

// WriteInt64 writes an int64 into the blob using little-endian.
func (b *Blob) WriteInt64(offset int, value int64) error {
	if err := b.checkRange(offset, 8); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(b.Data[offset:], uint64(value))
	return nil
}

// ReadFloat64 reads a float64 from the blob using little-endian.
func (b *Blob) ReadFloat64(offset int) (float64, error) {
	if err := b.checkRange(offset, 8); err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint64(b.Data[offset:])
	return math.Float64frombits(bits), nil
}

// WriteFloat64 writes a float64 into the blob using little-endian.
func (b *Blob) WriteFloat64(offset int, value float64) error {
	if err := b.checkRange(offset, 8); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(b.Data[offset:], math.Float64bits(value))
	return nil
}

func (b *Blob) checkRange(offset, size int) error {
	if b == nil {
		return errors.New("blob is nil")
	}
	if size < 0 || offset < 0 {
		return fmt.Errorf("offset out of range: %d", offset)
	}
	if size > len(b.Data) || offset > len(b.Data)-size {
		return fmt.Errorf("offset out of range: %d", offset)
	}
	return nil
}
