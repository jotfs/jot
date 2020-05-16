package iotafs

import (
	"encoding/binary"
	"io"

	"github.com/iotafs/iotafs-go/internal/sum"
)

const packfileObject uint8 = 1

// packfileBuilder is used to build a packfile object.
type packfileBuilder struct {
	w    *countingWriter
	hash *sum.Hash
}

// newPackfileBuilder creates a new packfileBuilder
func newPackfileBuilder(w io.Writer) (*packfileBuilder, error) {
	hash, err := sum.New()
	if err != nil {
		return nil, err
	}
	// Send everything writen to the packfile through the hash function
	w = io.MultiWriter(w, hash)
	wr := &countingWriter{w, 0}

	// Write the object type
	if _, err := wr.Write([]byte{packfileObject}); err != nil {
		return nil, err
	}

	b := packfileBuilder{wr, hash}
	return &b, nil
}

// append writes a chunk of data to the packfile.
func (b *packfileBuilder) append(data []byte, sum sum.Sum, mode compressMode) error {
	block := makeBlock(data, sum, mode)
	if _, err := b.w.Write(block); err != nil {
		return err
	}
	return nil
}

// size returns the current number of bytes written to the packfile.
func (b *packfileBuilder) size() uint64 {
	return b.w.size
}

// sum returns the checksum of all data written to the packfile so far.
func (b *packfileBuilder) sum() sum.Sum {
	return b.hash.Sum()
}

// makeBlock creates a packfile block in its binary format. The data should not be
// compressed beforehand.
func makeBlock(data []byte, s sum.Sum, mode compressMode) []byte {

	// Reserve the first 8 bytes for the size of the compressed data
	block := make([]byte, 8)
	block = append(block, mode.asUint8())
	block = append(block, s[:]...)
	before := len(block)
	block = mode.compress(block, data)

	// Set the size of the compressed data
	size := uint64(len(block) - before)
	binary.LittleEndian.PutUint64(block[:8], size)

	return block
}

type countingWriter struct {
	w    io.Writer
	size uint64
}

func (w *countingWriter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.size += uint64(n)
	return n, err
}
