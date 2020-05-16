package iotafs

import (
	"fmt"

	"github.com/valyala/gozstd"
)

type compressMode uint8

// Data compression modes
const (
	CompressNone compressMode = 0 // No compression
	CompressZStd compressMode = 1 // Zstandard compression
)

// asUint8 converts a compression mode to a uint8.
func (m compressMode) asUint8() uint8 {
	return uint8(m)
}

// fromUint8 converts a uint8 to a compression mode. Returns an error if the value
// is an unknown mode.
func fromUint8(v uint8) (compressMode, error) {
	if v <= 1 {
		return compressMode(v), nil
	}
	return 0, fmt.Errorf("invalid compression mode %d", v)
}

// Compress compresses src, appends it to dst, and returns the updated dst slice.
func (m compressMode) compress(dst []byte, src []byte) []byte {
	switch m {
	case CompressNone:
		dst = append(dst, src...)
		return dst
	case CompressZStd:
		dst = gozstd.Compress(dst, src)
		return dst
	default:
		panic("not implemented")
	}
}
