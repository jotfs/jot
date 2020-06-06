package jot

import (
	"fmt"
	"io"

	kzstd "github.com/klauspost/compress/zstd"
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

// compressModeFromUint8 converts a uint8 to a compression mode. Returns an error if the value
// is an unknown mode.
func compressModeFromUint8(v uint8) (CompressMode, error) {
	if v <= 1 {
		return CompressMode(v), nil
	}
	return 0, fmt.Errorf("invalid compression mode %d", v)
}

func (m CompressMode) compress(src []byte, dst []byte) ([]byte, error) {
	switch m {
	case CompressNone:
		dst = append(dst, src...)
		return dst, nil
	case CompressZstd:
		w, err := kzstd.NewWriter(nil)
		if err != nil {
			return nil, err
		}
		return w.EncodeAll(src, dst), nil
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
		r, err := kzstd.NewReader(src)
		if err != nil {
			return err
		}
		defer r.Close()
		_, err = io.Copy(dst, r)
		return err
	default:
		return fmt.Errorf("unknown compression mode %d", m)
	}
}
