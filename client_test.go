package jot

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// New
	_, err := New("https://localhost:8000", nil)
	assert.NoError(t, err)

	// New with options
	client, err := New("http://example.com", &Options{
		Compression: CompressNone,
		CacheDir:    "/var/log",
	})
	assert.NoError(t, err)
	assert.Equal(t, CompressNone, client.mode)
	assert.Equal(t, "/var/log", client.cacheDir)

	// Error if endpoint string is empty
	client, err = New("", nil)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestUpload(t *testing.T) {
	client := testClient(t)
	defer clearFiles(t, client)

	// Upload data
	data := randBytes(130*miB, 7456)
	_, err := client.Upload(bytes.NewReader(data), "data.txt")
	assert.NoError(t, err)

	// Upload empty data
	ctx := context.Background()
	_, err = client.UploadWithContext(ctx, bytes.NewReader(nil), "data2.txt", CompressNone)
	assert.NoError(t, err)
}

func TestDownload(t *testing.T) {
	client := testClient(t)
	defer clearFiles(t, client)

	// Upload data (none & zstd compression) and check download matches original data
	data1 := randBytes(10*miB, 2854)
	for _, mode := range []CompressMode{CompressNone, CompressZstd} {
		s1 := uploadTestFile(t, client, data1, "data1.txt", mode)

		msg := fmt.Sprintf("mode = %d", mode)
		var buf bytes.Buffer
		err := client.Download(s1, &buf)
		assert.NoError(t, err, msg)
		assert.Equal(t, data1, buf.Bytes(), msg)
	}

	// Download file which doesn't exist -- should get ErrNotFound
	var buf bytes.Buffer
	err := client.Download(FileID{}, &buf)
	assert.Equal(t, ErrNotFound, err)

}

func TestList(t *testing.T) {
	client := testClient(t)
	defer clearFiles(t, client)

	data1 := randBytes(2*kiB, 6332)
	data2 := randBytes(12*miB, 83)
	data3 := randBytes(10*miB, 9832)
	s1 := uploadTestFile(t, client, data1, "data1.txt", CompressNone)
	s2 := uploadTestFile(t, client, data2, "files/data2.txt", CompressZstd)
	s3 := uploadTestFile(t, client, data3, "files/data3.txt", CompressNone)

	// We don't know the CreatedAt value here so leave it at zero for each
	info1 := FileInfo{Name: "/data1.txt", Size: uint64(len(data1)), FileID: s1}
	info2 := FileInfo{Name: "/files/data2.txt", Size: uint64(len(data2)), FileID: s2}
	info3 := FileInfo{Name: "/files/data3.txt", Size: uint64(len(data3)), FileID: s3}

	runIt := func(it FileIterator) ([]FileInfo, error) {
		files := make([]FileInfo, 0)
		for {
			file, err := it.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			// Zero the CreatedAt field so we can match it with the expected values
			file.CreatedAt = time.Time{}
			files = append(files, file)
		}
		return files, nil
	}

	// List
	it := client.List("/", nil)
	files, err := runIt(it)
	assert.NoError(t, err)
	assert.Equal(t, []FileInfo{info3, info2, info1}, files)

	// List ascending
	it = client.List("/", &ListOpts{Ascending: true})
	files, err = runIt(it)
	assert.NoError(t, err)
	assert.Equal(t, []FileInfo{info1, info2, info3}, files)

	// List filter
	it = client.List("/", &ListOpts{Exclude: "/files/*", Include: "/files/data2.txt"})
	files, err = runIt(it)
	assert.NoError(t, err)
	assert.Equal(t, []FileInfo{info2, info1}, files)

	// List -- no matches
	it = client.List("/invalid/", nil)
	files, err = runIt(it)
	assert.NoError(t, err)
	assert.Equal(t, []FileInfo{}, files)

	// Head
	it = client.Head("data1.txt", nil)
	files, err = runIt(it)
	assert.NoError(t, err)
	assert.Equal(t, []FileInfo{info1}, files)

	// Head -- no matches
	it = client.Head("does-not-exist", nil)
	files, err = runIt(it)
	assert.NoError(t, err)
	assert.Equal(t, []FileInfo{}, files)
}

func TestCopy(t *testing.T) {
	client := testClient(t)
	defer clearFiles(t, client)

	data1 := randBytes(13*miB, 5343)
	s1 := uploadTestFile(t, client, data1, "data1.txt", CompressNone)

	// Copy
	s2, err := client.Copy(s1, "data/data2.txt")
	assert.NoError(t, err)

	// Should be able to get file
	var buf bytes.Buffer
	err = client.Download(s2, &buf)
	assert.NoError(t, err)
	assert.Equal(t, data1, buf.Bytes())

	// Copy does not exist
	_, err = client.Copy(FileID{}, "abc")
	assert.Equal(t, ErrNotFound, err)
}

func TestMergeErrors(t *testing.T) {
	err1 := errors.New("1")
	err2 := errors.New("2")

	tests := []struct {
		e      error
		minor  error
		output error
	}{
		{err1, nil, err1},
		{nil, err2, err2},
		{nil, nil, nil},
	}

	for i, test := range tests {
		assert.Equal(t, test.output, mergeErrors(test.e, test.minor), i)
	}

	// Check wrapping
	err := mergeErrors(err1, err2)
	assert.Equal(t, err1, errors.Unwrap(err))
}

func uploadTestFile(t *testing.T, client *Client, data []byte, name string, mode CompressMode) FileID {
	ctx := context.Background()
	s, err := client.UploadWithContext(ctx, bytes.NewReader(data), name, mode)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func clearFiles(t *testing.T, client *Client) {
	it := client.List("/", nil)
	for {
		file, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if err = client.Delete(file.FileID); err != nil {
			t.Fatal(err)
		}
	}
}

func testClient(t *testing.T) *Client {
	client, err := New("http://localhost:6777", nil)
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func randBytes(n int, seed int64) []byte {
	b := make([]byte, n)
	rand.Seed(seed)
	rand.Read(b)
	return b
}

func BenchmarkUpload(b *testing.B) {
	client, err := New("http://localhost:6777", nil)
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	f, err := os.Open("random.txt")
	if err != nil {
		b.Fatal(err)
	}
	_, err = client.UploadWithContext(ctx, f, "random.txt", CompressNone)
	if err != nil {
		b.Fatal(err)
	}
}
