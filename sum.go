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

// computeSum returns the checksum of a byte slice.
func computeSum(data []byte) FileID {
	h := blake3.New()
	h.Write(data)
	var s FileID
	copy(s[:], h.Sum(nil))
	return s
}

// asHex returns the hex-encoded representation of s.
func (s FileID) asHex() string {
	return hex.EncodeToString(s[:])
}

// sumFromHex converts a hex encoded string to a FileID.
func sumFromHex(s string) (FileID, error) {
	b, err := hex.DecodeString(s)
	if err != nil {
		return FileID{}, err
	}
	return UnmarshalFileID(b)
}

// sumHash computes a checksum. Implements the `io.Writer` interface.
type sumHash struct {
	h hash.Hash
}

// New returns a new Hash.
func newHash() *sumHash {
	h := blake3.New()
	return &sumHash{h}
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

// Marshal serializes a FileID to a byte slice.
func (s FileID) Marshal() []byte {
	b := make([]byte, sumSize)
	copy(b, s[:])
	return b
}

// UnmarshalFileID converts a byte slice to a FileID.
func UnmarshalFileID(b []byte) (FileID, error) {
	if len(b) != sumSize {
		return FileID{}, fmt.Errorf("length must be %d not %d", sumSize, len(b))
	}
	var s FileID
	copy(s[:], b)
	return s, nil
}
