package iotafs

import (
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/zeebo/blake3"
)

// sumSize is the byte-size of a checksum
const sumSize = 32

// FileID represents a unique ID for a file on an IotaFS server.
type FileID [sumSize]byte

// sumFromBytes converts a byte slice to a Sum. Its length must be sum.Size bytes.
func sumFromBytes(b []byte) (FileID, error) {
	if len(b) != sumSize {
		return FileID{}, fmt.Errorf("length must be %d not %d", sumSize, len(b))
	}
	var s FileID
	copy(s[:], b)
	return s, nil
}

// computeSum returns the checksum of a byte slice.
func computeSum(data []byte) FileID {
	h := blake3.New()
	h.Write(data)
	var s FileID
	copy(s[:], h.Sum(nil))
	return s
}

// AsHex returns the hex-encoded representation of s.
func (s FileID) AsHex() string {
	return hex.EncodeToString(s[:])
}

// sumFromHex converts a hex encoded string to a Sum.
func sumFromHex(s string) (FileID, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return FileID{}, err
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
func (h *sumHash) Sum() FileID {
	b := h.h.Sum(nil)
	var s FileID
	copy(s[:], b)
	return s
}
