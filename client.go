package iotafs

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/iotafs/fastcdc-go"
	"github.com/twitchtv/twirp"

	pb "github.com/iotafs/iotafs-go/internal/protos/upload"
	"github.com/iotafs/iotafs-go/internal/sum"
)

// ErrNotFound is returned when a file with a given ID cannot be found on the remote.
var ErrNotFound = errors.New("not found")

var networkError = errors.New("unable to connect to host")

const (
	kiB             = 1024
	miB             = 1024 * kiB
	maxPackfileSize = 128 * miB
)

var defaultCDCOpts = fastcdc.Options{
	MinSize:    256 * kiB,
	MaxSize:    8 * miB,
	NormalSize: 1 * miB,
	SmallBits:  22,
	LargeBits:  18,
}

// Client implements methods to interact with an IotaFS server.
type Client struct {
	host     url.URL
	hclient  *http.Client
	iclient  pb.IotaFS
	cacheDir string
}

// New returns a new Client. If applicable, host should include the port number, for
// example, "example.com:6776".
func New(host string) (*Client, error) {
	url, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("unable to parse host: %w", err)
	}

	hclient := &http.Client{}
	return &Client{
		host:    *url,
		hclient: hclient,
		iclient: pb.NewIotaFSProtobufClient(host, hclient),
	}, nil
}

// UploadWithContext uploads a new file with a given name.
func (c *Client) UploadWithContext(ctx context.Context, r io.Reader, dst string, mode CompressMode) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start uploading packfiles on a background goroutine
	packFiles := make(chan string)
	done := make(chan error)
	go c.uploadPackfiles(ctx, packFiles, done)

	// Create a directory to cache data during the upload
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	dir := filepath.Join(c.cacheDir, id.String())
	if err := os.Mkdir(dir, 0744); err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	packer, err := newPacker(dir, packFiles)

	chunker, err := fastcdc.NewChunker(r, defaultCDCOpts)
	if err != nil {
		return err
	}

	fileSums := make([][]byte, 0)
	cache := newCache(5 * defaultCDCOpts.MaxSize) // TODO: make cache size parameter

	// Add new chunks from the file to one or more packfiles, and send the resulting
	// packfiles to the uploader
	err = func() error {
		defer close(packFiles)
		eof := false
		for i := 0; ; i++ {
			chunk, err := chunker.Next()
			if err == io.EOF {
				eof = true
			} else if err != nil {
				return err
			}

			if eof || !cache.hasCapacity(chunk.Data) {
				// Check which chunks in the cache need to be added to a packfile
				resp, err := c.iclient.ChunksExist(ctx, &pb.ChunksExistRequest{Sums: sumsToBytes(cache.sums)})
				if err != nil {
					return err
				}
				for j, sum := range cache.sums {
					if resp.Exists[j] {
						continue
					}
					data := cache.getChunk(sum)
					if err := packer.addChunk(data, sum, mode); err != nil {
						return err
					}
				}
				cache.clear()
			}
			if eof {
				break
			}

			s := sum.Compute(chunk.Data)
			if err := cache.saveChunk(chunk.Data, s); err != nil {
				return err
			}
			fileSums = append(fileSums, s[:])
		}

		return packer.flush() // send any remaining data to the uploader
	}()
	if err != nil {
		cancel()
		return err
	}

	// Wait for the uploader to complete.
	err = <-done
	if err != nil {
		return err
	}

	// Create the file
	_, err = c.iclient.CreateFile(ctx, &pb.File{Name: dst, Sums: fileSums})
	if err != nil {
		return err
	}

	return nil
}

// location stores the size and offset of a chunk of data in the cache.
type location struct {
	size   int
	offset int
}

// cache is fixed-size buffer for temporarily storing chunk data.
type cache struct {
	cursor int
	buf    []byte
	sums   []sum.Sum
	chunks map[sum.Sum]location
}

// newCache creates a new cache with a given buffer size.
func newCache(size int) *cache {
	return &cache{0, make([]byte, size), make([]sum.Sum, 0), make(map[sum.Sum]location)}
}

// hasCapacity returns true if the cache has enough room to store data.
func (c *cache) hasCapacity(data []byte) bool {
	if len(c.buf)-c.cursor < len(data) {
		return false
	}
	return true
}

// saveChunk copies chunk data with a given checksum to the cache.
func (c *cache) saveChunk(data []byte, s sum.Sum) error {
	if !c.hasCapacity(data) {
		return fmt.Errorf("unable to save chunk of size %d into cache", len(data))
	}
	if _, ok := c.chunks[s]; ok {
		return nil
	}

	n := len(data)
	copy(c.buf[c.cursor:c.cursor+n], data)
	c.chunks[s] = location{size: n, offset: c.cursor}
	c.cursor += n
	c.sums = append(c.sums, s)

	return nil
}

// getChunk returns chunk data from the cache with a given checksum. Will panic if a
// chunk with the provided checksum has not already been saved.
func (c *cache) getChunk(s sum.Sum) []byte {
	loc := c.chunks[s]
	return c.buf[loc.offset : loc.offset+loc.size]
}

// clear removes all data from the cache.
func (c *cache) clear() {
	c.cursor = 0
	c.chunks = make(map[sum.Sum]location)
	c.sums = c.sums[:0]
}

// uploadPackfiles listens for packfiles on a channel and uploads them. It signals
// completion (a nil error) or failure on the done channel.
func (c *Client) uploadPackfiles(ctx context.Context, files <-chan string, done chan<- error) {
	u := c.host
	u.Path = path.Join(u.Path, "packfile")
	url := u.String()

	upload := func(name string) error {
		// Get the packfile checksum from its file name
		s, err := sum.FromHex(filepath.Base(name))
		if err != nil {
			return err
		}

		// Get the packfile from the cache
		f, err := os.Open(name)
		if err != nil {
			return err
		}
		defer f.Close()
		info, err := f.Stat()
		if err != nil {
			return err
		}

		// Construct the request
		req, err := http.NewRequestWithContext(ctx, "POST", url, f)
		if err != nil {
			return err
		}
		req.Header.Set("x-iota-checksum", base64.StdEncoding.EncodeToString(s[:]))
		req.ContentLength = info.Size()

		// Upload the file
		resp, err := c.hclient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			b, _ := ioutil.ReadAll(resp.Body)
			msg := string(b)
			err = fmt.Errorf("chunk upload failed (%d): %s", resp.StatusCode, msg)
			return err
		}

		return nil
	}

	for name := range files {
		err := upload(name)
		err = mergeErrors(err, os.Remove(name))
		if err != nil {
			done <- err
		}
	}
	done <- nil
}

func sumsToBytes(sums []sum.Sum) [][]byte {
	b := make([][]byte, len(sums))
	for i, s := range sums {
		b[i] = s[:]
	}
	return b
}

// packer adds chunks to a packfile. When its current packfile is at capacity, it sends
// it to the uploader on the packFiles channel, and accepts a new one.
type packer struct {
	dir       string
	packFiles chan<- string

	f       *os.File
	builder *packfileBuilder
}

func newPacker(dir string, packFiles chan<- string) (*packer, error) {
	p := &packer{dir, packFiles, nil, nil}
	err := p.initBuilder()
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (p *packer) initBuilder() error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	name := filepath.Join(p.dir, id.String())
	f, err := os.OpenFile(name, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	builder, err := newPackfileBuilder(f)
	if err != nil {
		return err
	}
	p.f = f
	p.builder = builder
	return nil
}

// flush closes the current builder and sends the packfile to the uploader.
func (p *packer) flush() error {
	if err := p.f.Close(); err != nil {
		return err
	}
	if p.builder.size() == 0 {
		// Do nothing if the packfile is empty.
		return os.Remove(p.f.Name())
	}
	packSum := p.builder.sum()
	packName := filepath.Join(p.dir, packSum.AsHex())
	if err := os.Rename(p.f.Name(), packName); err != nil {
		return err
	}
	p.packFiles <- packName

	return nil
}

// addChunk adds a chunk to a packfile owned by the packer.
func (p *packer) addChunk(data []byte, sum sum.Sum, mode CompressMode) error {
	if p.builder.size()+uint64(len(data)) > maxPackfileSize {
		if err := p.flush(); err != nil {
			return err
		}
		if err := p.initBuilder(); err != nil {
			return err
		}
	}
	return p.builder.append(data, sum, mode)
}

type FileInfo struct {
	Name      string
	CreatedAt time.Time
	Size      uint64
	Sum       sum.Sum
}

// FileIterator is an iterator over a stream of FileInfo returned by List and Head.
type FileIterator interface {

	// Next returns the next FileInfo from the iterator. It returns io.EOF when the end
	// of the iterator is reached. The FileInfo is always invalid if the error is not nil.
	Next() (FileInfo, error)
}

// IteratorOpts specify the options for an iterator.
type IteratorOpts struct {
	// Limit is the maximum number of values to return from the iterator. Unlimited if
	// unspecified.
	Limit uint64

	// BatchSize is the maximum number of values to retrieve at a time from the remote.
	// Defaults to 1000 if unspecified.
	BatchSize uint64

	// Ascending, if set to true, returns values from the iterator in chronological
	// order. False by default, in which case values are returned in reverse-chronological
	// order
	Ascending bool
}

func (c *Client) List(prefix string, opts *IteratorOpts) FileIterator {
	itOpts := defaultIteratorOpts(opts)
	return &listIterator{opts: itOpts, prefix: prefix, iclient: c.iclient}
}

type listIterator struct {
	opts    IteratorOpts
	prefix  string
	iclient pb.IotaFS

	nextPageToken int64
	values        []*pb.FileInfo
	cursor        int
	count         uint64
}

func (it *listIterator) Next() (FileInfo, error) {
	if it.opts.Limit != 0 && it.count == it.opts.Limit {
		return FileInfo{}, io.EOF
	}
	if it.cursor == len(it.values) {
		if it.nextPageToken == -1 {
			return FileInfo{}, io.EOF
		}

		// Get a new batch
		ctx := context.Background()
		resp, err := it.iclient.List(ctx, &pb.ListRequest{
			Prefix:        it.prefix,
			Limit:         it.opts.BatchSize,
			NextPageToken: it.nextPageToken,
		})
		if isNetworkError(err) {
			return FileInfo{}, networkError
		}
		if err != nil {
			return FileInfo{}, err
		}

		it.values = resp.Info
		it.nextPageToken = resp.NextPageToken
		it.cursor = 0
	}
	if len(it.values) == 0 {
		return FileInfo{}, io.EOF
	}

	v := it.values[it.cursor]
	s, err := sum.FromBytes(v.Sum)
	if err != nil {
		return FileInfo{}, err
	}
	info := FileInfo{Name: v.Name, CreatedAt: time.Unix(0, v.CreatedAt), Size: v.Size, Sum: s}

	it.cursor++
	it.count++

	return info, nil
}

type headIterator struct {
	opts    IteratorOpts
	name    string
	iclient pb.IotaFS

	nextPageToken int64
	values        []*pb.FileInfo
	cursor        int
	count         uint64
}

func (it *headIterator) Next() (FileInfo, error) {
	if it.opts.Limit != 0 && it.count == it.opts.Limit {
		return FileInfo{}, io.EOF
	}
	if it.cursor == len(it.values) {
		if it.nextPageToken == -1 {
			return FileInfo{}, io.EOF
		}

		// Get a new batch
		ctx := context.Background()
		resp, err := it.iclient.Head(ctx, &pb.HeadRequest{
			Name:          it.name,
			Limit:         it.opts.BatchSize,
			NextPageToken: it.nextPageToken,
		})
		if isNetworkError(err) {
			return FileInfo{}, networkError
		}
		if err != nil {
			return FileInfo{}, err
		}

		it.values = resp.Info
		it.nextPageToken = resp.NextPageToken
		it.cursor = 0
	}
	if len(it.values) == 0 {
		return FileInfo{}, io.EOF
	}

	v := it.values[it.cursor]
	s, err := sum.FromBytes(v.Sum)
	if err != nil {
		return FileInfo{}, err
	}
	info := FileInfo{Name: v.Name, CreatedAt: time.Unix(0, v.CreatedAt), Size: v.Size, Sum: s}

	it.cursor++
	it.count++

	return info, nil
}

func defaultIteratorOpts(opts *IteratorOpts) IteratorOpts {
	var itOpts IteratorOpts
	itOpts.BatchSize = 1000
	if opts != nil {
		itOpts.Ascending = opts.Ascending
		itOpts.Limit = opts.Limit
		if opts.BatchSize != 0 {
			itOpts.BatchSize = opts.BatchSize
		}
	}
	return itOpts
}

func (c *Client) Head(name string, opts *IteratorOpts) FileIterator {
	itOpts := defaultIteratorOpts(opts)
	return &headIterator{opts: itOpts, name: name, iclient: c.iclient}
}

func min(a, b uint64) uint64 {
	if a < b {
		return a
	}
	return b
}

// Download retrieves a file and writes it to dst. Returns iotafs.ErrNotFound if the
// file does not exist.
func (c *Client) Download(file sum.Sum, dst io.Writer) error {
	ctx := context.Background()
	resp, err := c.iclient.Download(ctx, &pb.FileID{Sum: file[:]})
	if e, ok := err.(twirp.Error); ok && e.Code() == twirp.NotFound {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Download the data for each packfile section using the provided URLs, and use
	// the data to construct the original file.
	for i, s := range resp.Sections {
		err := c.downloadSection(dst, s)
		if err != nil {
			return fmt.Errorf("section %d: %w", i, err)
		}
	}

	return nil
}

func (c *Client) downloadSection(dst io.Writer, s *pb.Section) error {
	// Create temp file to hold the section data
	tmp, err := ioutil.TempFile("", "")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	// Construct a request to get the section data
	req, err := http.NewRequest("GET", s.Url, nil)
	if err != nil {
		return err
	}
	length := s.RangeEnd - s.RangeStart
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", s.RangeStart, s.RangeEnd))

	resp, err := c.hclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if n, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	} else if n != int64(length) {
		return fmt.Errorf("expected %d bytes but only received %d", length, n)
	}

	// Add the data for each chunk to the file, ensuring the checksum matches
	for _, chunk := range s.Chunks {
		_, err = tmp.Seek(int64(chunk.BlockOffset), io.SeekStart)
		if err != nil {
			return err
		}

		block, err := readBlock(tmp)
		if err != nil {
			return fmt.Errorf("reading block: %w", err)
		}

		h, err := sum.New()
		if err != nil {
			return err
		}
		w := io.MultiWriter(dst, h)
		if err := block.Mode.decompressStream(w, bytes.NewReader(block.Data)); err != nil {
			return fmt.Errorf("decompression block: %w", err)
		}

		s := h.Sum()
		if s != block.Sum {
			return fmt.Errorf("actual chunk checksum %x does not match block sum %x", s, block.Sum)
		}
	}

	return nil
}

func mergeErrors(e error, minor error) error {
	if e == nil && minor == nil {
		return nil
	}
	if e == nil {
		return minor
	}
	if minor == nil {
		return e
	}
	return fmt.Errorf("%w; %v", e, minor)
}

// Copy copies a file from one IotaFS location to another IotaFS location and returns
// the ID of the new file. Returns iotafs.ErrNotFound if the source file does not exist.
func (c *Client) Copy(src sum.Sum, dst string) (sum.Sum, error) {
	ctx := context.Background()
	fileID, err := c.iclient.Copy(ctx, &pb.CopyRequest{SrcId: src[:], Dst: dst})
	if e, ok := err.(twirp.Error); ok && e.Code() == twirp.NotFound {
		return sum.Sum{}, ErrNotFound
	}
	if err != nil {
		return sum.Sum{}, err
	}
	s, err := sum.FromBytes(fileID.Sum)
	if err != nil {
		return sum.Sum{}, err
	}
	return s, nil
}

// Delete deletes a file. Returns iotafs.ErrNotFound if the file does not exist.
func (c *Client) Delete(file sum.Sum) error {
	ctx := context.Background()
	_, err := c.iclient.Delete(ctx, &pb.FileID{Sum: file[:]})
	if e, ok := err.(twirp.Error); ok && e.Code() == twirp.NotFound {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return err
}

// isNetworkError checks if an error is returned because the client cannot connect to
// the server.
func isNetworkError(e error) bool {
	if e == nil {
		return false
	}
	if strings.Contains(e.Error(), "failed to do request") {
		return true
	}
	return false
}
