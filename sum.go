package iotafs

import (
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/zeebo/blake3"
)

// sumSize is the byte-size of a checksum
const sumSize = 32

// Sum stores a checksum
type Sum [sumSize]byte

// sumFromBytes converts a byte slice to a Sum. Its length must be sum.Size bytes.
func sumFromBytes(b []byte) (Sum, error) {
	if len(b) != sumSize {
		return Sum{}, fmt.Errorf("length must be %d not %d", sumSize, len(b))
	}
	var s Sum
	copy(s[:], b)
	return s, nil
}

// computeSum returns the checksum of a byte slice.
func computeSum(data []byte) Sum {
	h := blake3.New()
	h.Write(data)
	var s Sum
	copy(s[:], h.Sum(nil))
	return s
}

// AsHex returns the hex-encoded representation of s.
func (s Sum) AsHex() string {
	return hex.EncodeToString(s[:])
}

// sumFromHex converts a hex encoded string to a Sum.
func sumFromHex(s string) (Sum, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return Sum{}, err
	}
	return sumFromBytes(b)
}

// sumHash computes a checksum. Implements the `io.Writer` interface.
type sumHash struct {
	h hash.Hash
}

// New returns a new Hash.
func newHash() (*sumHash, error) {
	h := blake3.New()
	return &sumHash{h}, nil
}

// Write writes a byte slice to the hash function.
func (h *sumHash) Write(p []byte) (int, error) {
	return h.h.Write(p)
}

// Sum returns the current checksum of a Hash.
func (h *sumHash) Sum() Sum {
	b := h.h.Sum(nil)
	var s Sum
	copy(s[:], b)
	return s
}
