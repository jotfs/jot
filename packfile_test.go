package jot

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPackfileBuilder(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	builder, err := newPackfileBuilder(buf)
	assert.NoError(t, err)
	assert.NotNil(t, builder)

	aSum := computeSum(a)
	err = builder.append(a, aSum, CompressNone)
	assert.NoError(t, err)

	bSum := computeSum(b)
	err = builder.append(b, bSum, CompressZstd)
	assert.NoError(t, err)

	assert.NotZero(t, builder.size())

	// Deserialize the packfile
	// First byte should be object type
	otype, err := buf.ReadByte()
	assert.NoError(t, err)
	assert.Equal(t, packfileObject, otype)

	// Read first block
	b1, err := readBlock(buf)
	assert.NoError(t, err)
	assert.Equal(t, CompressNone, b1.Mode)
	assert.Equal(t, aSum, b1.Sum)

	// Read second block
	b2, err := readBlock(buf)
	assert.NoError(t, err)
	assert.Equal(t, CompressZstd, b2.Mode)
	assert.Equal(t, bSum, b2.Sum)

	// Decompress data and check that it matches the original
	var b1d bytes.Buffer
	err = b1.Mode.decompressStream(&b1d, bytes.NewReader(b1.Data))
	assert.NoError(t, err)
	assert.Equal(t, a, b1d.Bytes())

	var b2d bytes.Buffer
	err = b2.Mode.decompressStream(&b2d, bytes.NewReader(b2.Data))
	assert.NoError(t, err)
	assert.Equal(t, b, b2d.Bytes())
}

var a = []byte(`A celebrated tenor had sung in Italian, and a notorious contralto had sung 
 in jazz, and between the numbers people were doing "stunts." all over the garden, while 
happy, vacuous bursts of laughter rose toward the summer sky. A pair of stage twins, who 
turned out to be the girls in yellow, did a baby act in costume, and champagne was served 
in glasses bigger than finger-bowls. The moon had risen higher, and floating in the Sound 
was a triangle of silver scales, trembling a little to the stiff, tinny drip of the 
banjoes on the lawn.`)

var b = []byte(`And as I sat there brooding on the old, unknown world, I thought of Gatsby’s 
wonder when he first picked out the green light at the end of Daisy’s dock. He had come 
a long way to this blue lawn, and his dream must have seemed so close that he could 
hardly fail to grasp it. He did not know that it was already behind him, somewhere back 
in that vast obscurity beyond the city, where the dark fields of the republic rolled on 
under the night. Gatsby believed in the green light, the orgastic future that year by 
year recedes before us. It eluded us then, but that’s no matter -- tomorrow we will run 
faster, stretch out our arms farther... And one fine morning -- So we beat on, boats 
against the current, borne back ceaselessly into the past.`)
