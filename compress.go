package iotafs

import (
	"fmt"
	"io"

	"github.com/DataDog/zstd"
)

// CompressMode is a type alias for compression modes.
type CompressMode uint8

// Data compression modes
const (
	CompressZstd CompressMode = 0 // Zstandard compression
	CompressNone CompressMode = 1 // No compression
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

// // Compress compresses src, appends it to dst, and returns the updated dst slice.
// func (m compressMode) compress(dst []byte, src []byte) []byte {
// 	switch m {
// 	case CompressNone:
// 		dst = append(dst, src...)
// 		return dst
// 	case CompressZstd:
// 		// zstd.Com

// 		// dst = gozstd.Compress(dst, src)
// 		return dst
// 	default:
// 		panic("not implemented")
// 	}
// }
