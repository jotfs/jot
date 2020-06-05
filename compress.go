package jot

import (
	"fmt"
	"io"

	"github.com/DataDog/zstd"
)

// CompressMode represents the algorithm used to compress data.
type CompressMode uint8

// Data compression modes
const (
	// Zstandard compression
	CompressZstd CompressMode = 0
	// No compression
	CompressNone CompressMode = 1
)

// asUint8 converts a compression mode to a uint8.
func (m CompressMode) asUint8() uint8 {
	return uint8(m)
}

// fromUint8 converts a uint8 to a compression mode. Returns an error if the value
// is an unknown mode.
func fromUint8(v uint8) (CompressMode, error) {
	if v <= 1 {
		return CompressMode(v), nil
	}
	return 0, fmt.Errorf("invalid compression mode %d", v)
}

// compress compresses src, appends it to dst, and returns the updated dst slice.
func (m CompressMode) compress(src []byte) ([]byte, error) {
	switch m {
	case CompressNone:
		dst := make([]byte, len(src))
		copy(dst, src)
		return dst, nil
	case CompressZstd:
		return zstd.Compress(nil, src)
	default:
		panic("not implemented")
	}
}

// decompressStream decompresses data from src and writes it to dst.
func (m CompressMode) decompressStream(dst io.Writer, src io.Reader) error {
	switch m {
	case CompressNone:
		_, err := io.Copy(dst, src)
		return err
	case CompressZstd:
		r := zstd.NewReader(src)
		_, err := io.Copy(dst, r)
		cerr := r.Close()
		return mergeErrors(err, cerr)
	default:
		return fmt.Errorf("unknown compression mode %d", m)
	}
}
