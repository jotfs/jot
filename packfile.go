package iotafs

import (
	"encoding/binary"
	"fmt"
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

	b := packfileBuilder{wr, hash}
	return &b, nil
}

// append writes a chunk of data to the packfile.
func (b *packfileBuilder) append(data []byte, sum sum.Sum, mode compressMode) error {
	if b.size() == 0 {
		if _, err := b.w.Write([]byte{packfileObject}); err != nil {
			return fmt.Errorf("setting packfile object type: %w", err)
		}
	}
	block, err := makeBlock(data, sum, mode)
	if err != nil {
		return fmt.Errorf("appending chunk to packfile: %w", err)
	}
	if _, err := b.w.Write(block); err != nil {
		return fmt.Errorf("appending chunk to packfile: %w", err)
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
func makeBlock(data []byte, s sum.Sum, mode compressMode) ([]byte, error) {
	compressed, err := mode.compress(data)
	if err != nil {
		return nil, err
	}

	capacity := 8 + 1 + sum.Size + len(data)
	block := make([]byte, 8, capacity)

	binary.LittleEndian.PutUint64(block[:8], uint64(len(compressed)))
	block = append(block, mode.asUint8())
	block = append(block, s[:]...)
	block = append(block, compressed...)

	return block, nil
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
