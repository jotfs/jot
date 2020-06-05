package jot

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

	"github.com/jotfs/fastcdc-go"
	"github.com/rs/xid"
	"github.com/twitchtv/twirp"
	"golang.org/x/sync/errgroup"

	pb "github.com/jotfs/jot/internal/protos"
)

// ErrNotFound is returned when a file with a given ID cannot be found on the remote.
var ErrNotFound = errors.New("not found")

var errNetwork = errors.New("unable to connect to host")

const (
	kiB             = 1024
	miB             = 1024 * kiB
	maxPackfileSize = 128 * miB
)

// Client implements methods to interact with an JotFS server.
type Client struct {
	host     url.URL
	hclient  *http.Client
	iclient  pb.JotFS
	cacheDir string
	mode     CompressMode
	opts     *fastcdc.Options
}

// Options specifies optional configuration for a Client.
type Options struct {
	// Compression, if set, overrides the default compression mode, CompressZstd.
	Compression CompressMode

	// Timeout specifies the maximum time limit for HTTP requests made by the client.
	// It is unlimited by default.
	Timeout time.Duration

	// CacheDir specifies the directory the client may cache temporary data files.
	// os.TempDir() by default. The directory must already exist.
	CacheDir string
}

// New returns a new Client connecting to an JotFS server at the given endpoint URL.
// Optional configuration may be set with opts.
func New(endpoint string, opts *Options) (*Client, error) {
	url, err := url.ParseRequestURI(endpoint)
	if err != nil {
		return nil, fmt.Errorf("unable to parse endpoint: %v", err)
	}

	var timeout time.Duration
	mode := CompressZstd
	cacheDir := os.TempDir()
	if opts != nil {
		mode = opts.Compression
		timeout = opts.Timeout
		if opts.CacheDir != "" {
			cacheDir = opts.CacheDir
		}
	}

	hclient := &http.Client{Timeout: timeout}
	return &Client{
		host:     *url,
		hclient:  hclient,
		iclient:  pb.NewJotFSProtobufClient(endpoint, hclient),
		cacheDir: cacheDir,
		mode:     mode,
	}, nil
}

// Upload reads data from r and uploads it to the server, creating a new file version at
// dst. If a file already exists at the destination it will be kept if either:
//  1. File versioning is currently enabled on the server
//  2. File versioning was enabled when the existing version was uploaded
// Otherwise the latest version will be overwritten.
func (c *Client) Upload(r io.Reader, dst string) (FileID, error) {
	return c.UploadWithContext(context.Background(), r, dst, c.mode)
}

// UploadWithContext is the same as Upload with a user-supplied context and compression
// mode which overrides that set for the client.
func (c *Client) UploadWithContext(ctx context.Context, r io.Reader, dst string, mode CompressMode) (FileID, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Start uploading packfiles on a background goroutine
	var g errgroup.Group
	packFiles := make(chan string)
	g.Go(func() error {
		for name := range packFiles {
			err := c.uploadPackfile(ctx, name)
			err = mergeErrors(err, os.Remove(name))
			if err != nil {
				return err
			}
		}
		return nil
	})

	// Create a directory to cache data during the upload
	id := xid.New().String()
	dir := filepath.Join(c.cacheDir, "jot-"+id)
	if err := os.Mkdir(dir, 0744); err != nil {
		return FileID{}, err
	}
	defer os.RemoveAll(dir)

	packer, err := newPacker(dir, packFiles)

	// Get the chunker options from the server if we haven't done so already
	if c.opts == nil {
		opts, err := c.getChunkerOptions(ctx)
		if err != nil {
			return FileID{}, fmt.Errorf("retrieving chunker options: %w", err)
		}
		c.opts = opts
	}

	chunker, err := fastcdc.NewChunker(r, *c.opts)
	if err != nil {
		return FileID{}, err
	}

	fileSums := make([][]byte, 0)
	cache := newCache(5 * c.opts.MaxSize) // TODO: make cache size parameter

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
				req := &pb.ChunksExistRequest{Sums: sumsToBytes(cache.sums)}
				resp, err := c.iclient.ChunksExist(ctx, req)
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

			s := computeSum(chunk.Data)
			if err := cache.saveChunk(chunk.Data, s); err != nil {
				return err
			}
			fileSums = append(fileSums, s[:])
		}

		return packer.flush() // send any remaining data to the uploader
	}()
	if err != nil {
		cancel()
		return FileID{}, err
	}

	if err = g.Wait(); err != nil {
		return FileID{}, err
	}

	// Create the file
	info, err := c.iclient.CreateFile(ctx, &pb.File{Name: dst, Sums: fileSums})
	if err != nil {
		return FileID{}, err
	}
	sum, err := UnmarshalFileID(info.Sum)
	if err != nil {
		return FileID{}, err
	}

	return sum, nil
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
	sums   []FileID
	chunks map[FileID]location
}

// newCache creates a new cache with a given buffer size.
func newCache(size int) *cache {
	return &cache{0, make([]byte, size), make([]FileID, 0), make(map[FileID]location)}
}

// hasCapacity returns true if the cache has enough room to store data.
func (c *cache) hasCapacity(data []byte) bool {
	if len(c.buf)-c.cursor < len(data) {
		return false
	}
	return true
}

// saveChunk copies chunk data with a given checksum to the cache.
func (c *cache) saveChunk(data []byte, s FileID) error {
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
func (c *cache) getChunk(s FileID) []byte {
	loc := c.chunks[s]
	return c.buf[loc.offset : loc.offset+loc.size]
}

// clear removes all data from the cache.
func (c *cache) clear() {
	c.cursor = 0
	c.chunks = make(map[FileID]location)
	c.sums = c.sums[:0]
}

// uploadPackfile uploads a packfile to the server. The basename of the filepath
// must be the hex-encoded checksum of the file contents.
func (c *Client) uploadPackfile(ctx context.Context, fpath string) error {
	u := c.host
	u.Path = path.Join(u.Path, "packfile")
	url := u.String()

	// Get the packfile checksum from its file name
	s, err := sumFromHex(filepath.Base(fpath))
	if err != nil {
		return err
	}

	// Get the packfile from the cache
	f, err := os.Open(fpath)
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
	req.Header.Set("x-jotfs-checksum", base64.StdEncoding.EncodeToString(s[:]))
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

func sumsToBytes(sums []FileID) [][]byte {
	b := make([][]byte, len(sums))
	for i := range sums {
		b[i] = sums[i][:]
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
	id := xid.New().String()
	name := filepath.Join(p.dir, id)
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
	packName := filepath.Join(p.dir, packSum.asHex())
	if err := os.Rename(p.f.Name(), packName); err != nil {
		return err
	}
	p.packFiles <- packName

	return nil
}

// addChunk adds a chunk to a packfile owned by the packer.
func (p *packer) addChunk(data []byte, sum FileID, mode CompressMode) error {
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

// FileInfo stores metadata related to a file.
type FileInfo struct {
	Name      string
	CreatedAt time.Time
	Size      uint64
	FileID    FileID
}

// FileIterator is an iterator over a stream of FileInfo returned by the methods List and
// Head.
type FileIterator interface {

	// Next returns the next FileInfo from the iterator. It returns io.EOF when the end
	// of the iterator is reached. The FileInfo is always invalid if the error is not nil.
	Next() (FileInfo, error)
}

// ListOpts specify the options for the method List.
type ListOpts struct {
	// Exclude will exclude any files matching the glob pattern.
	Exclude string

	// Include forces inclusion of any files matched by the Exclude pattern. It is
	// ignored if Exclude is not provided.
	Include string

	// Limit is the maximum number of values to return from the iterator. Unlimited if
	// unspecified.
	Limit uint64

	// BatchSize is the maximum number of values to retrieve at a time from the remote.
	// Defaults to 1000.
	BatchSize uint64

	// Ascending, if set to true, returns values from the iterator in chronological
	// order. False by default, in which case values are returned in reverse-chronological
	// order
	Ascending bool
}

// List returns an iterator over all versions of all files with a name matching the
// given prefix. Files may be excluded / included by supplying a ListOpts struct.
func (c *Client) List(prefix string, opts *ListOpts) FileIterator {
	return c.ListWithContext(context.Background(), prefix, opts)
}

// ListWithContext is the same as List with a user supplied context.
func (c *Client) ListWithContext(ctx context.Context, prefix string, opts *ListOpts) FileIterator {
	var lopts ListOpts
	if opts != nil {
		lopts = *opts
	}
	if lopts.BatchSize == 0 {
		lopts.BatchSize = 1000
	}
	return &listIterator{ctx: ctx, opts: lopts, prefix: prefix, iclient: c.iclient}
}

type listIterator struct {
	ctx     context.Context
	opts    ListOpts
	prefix  string
	iclient pb.JotFS

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
		resp, err := it.iclient.List(it.ctx, &pb.ListRequest{
			Prefix:        it.prefix,
			Limit:         it.opts.BatchSize,
			NextPageToken: it.nextPageToken,
			Exclude:       it.opts.Exclude,
			Include:       it.opts.Include,
			Ascending:     it.opts.Ascending,
		})
		if isNetworkError(err) {
			return FileInfo{}, errNetwork
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
	s, err := UnmarshalFileID(v.Sum)
	if err != nil {
		return FileInfo{}, err
	}
	info := FileInfo{Name: v.Name, CreatedAt: time.Unix(0, v.CreatedAt).UTC(), Size: v.Size, FileID: s}

	it.cursor++
	it.count++

	return info, nil
}

// HeadOpts specify the options for the method Head.
type HeadOpts struct {
	// Limit is the maximum number of values to return from the iterator. Unlimited if
	// unspecified.
	Limit uint64

	// BatchSize is the maximum number of values to retrieve at a time from the remote.
	// Defaults to 1000.
	BatchSize uint64

	// Ascending, if set to true, returns values from the iterator in chronological
	// order. False by default, in which case values are returned in reverse-chronological
	// order
	Ascending bool
}

type headIterator struct {
	ctx     context.Context
	opts    HeadOpts
	name    string
	iclient pb.JotFS

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
		resp, err := it.iclient.Head(it.ctx, &pb.HeadRequest{
			Name:          it.name,
			Limit:         it.opts.BatchSize,
			NextPageToken: it.nextPageToken,
		})
		if isNetworkError(err) {
			return FileInfo{}, errNetwork
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
	s, err := UnmarshalFileID(v.Sum)
	if err != nil {
		return FileInfo{}, err
	}
	info := FileInfo{Name: v.Name, CreatedAt: time.Unix(0, v.CreatedAt).UTC(), Size: v.Size, FileID: s}

	it.cursor++
	it.count++

	return info, nil
}

// Head returns an iterator over the versions of a file with a given name.
func (c *Client) Head(name string, opts *HeadOpts) FileIterator {
	return c.HeadWithContext(context.Background(), name, opts)
}

// HeadWithContext is the same as Head with a user supplied context.
func (c *Client) HeadWithContext(ctx context.Context, name string, opts *HeadOpts) FileIterator {
	var hopts HeadOpts
	if opts != nil {
		hopts = *opts
	}
	if hopts.BatchSize == 0 {
		hopts.BatchSize = 1000
	}
	return &headIterator{ctx: ctx, opts: hopts, name: name, iclient: c.iclient}
}

// Download retrieves a file and writes it to w. Returns jotfs.ErrNotFound if the file
// does not exist.
func (c *Client) Download(file FileID, w io.Writer) error {
	return c.DownloadWithContext(context.Background(), file, w)
}

// DownloadWithContext is the same as Download with a user supplied context.
func (c *Client) DownloadWithContext(ctx context.Context, file FileID, w io.Writer) error {
	resp, err := c.iclient.Download(ctx, &pb.FileID{Sum: file[:]})
	if e, ok := err.(twirp.Error); ok && e.Code() == twirp.NotFound {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	// Download the data for each packfile section using the provided URLs, and use
	// the data to construct the original file.
	for i := range resp.Sections {
		section := resp.Sections[i]
		err := c.downloadSection(w, section)
		if err != nil {
			return fmt.Errorf("section %d %s: %w", i, section.Url, err)
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
	length := s.RangeEnd - s.RangeStart + 1
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", s.RangeStart, s.RangeEnd))

	resp, err := c.hclient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if n, err := io.Copy(tmp, resp.Body); err != nil {
		return err
	} else if n != int64(length) {
		return fmt.Errorf("expected %d bytes but received %d", length, n)
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

		hash := newHash()
		w := io.MultiWriter(dst, hash)
		if err := block.Mode.decompressStream(w, bytes.NewReader(block.Data)); err != nil {
			return fmt.Errorf("decompressing block: %w", err)
		}

		s := hash.Sum()
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

// Copy makes a copy of a file from src to dst and returns the ID of the new file.
// Returns jotfs.ErrNotFound if the src does not exist.
func (c *Client) Copy(src FileID, dst string) (FileID, error) {
	return c.CopyWithContext(context.Background(), src, dst)
}

// CopyWithContext is the same as Copy with a user supplied context.
func (c *Client) CopyWithContext(ctx context.Context, src FileID, dst string) (FileID, error) {
	fileID, err := c.iclient.Copy(ctx, &pb.CopyRequest{SrcId: src[:], Dst: dst})
	if e, ok := err.(twirp.Error); ok && e.Code() == twirp.NotFound {
		return FileID{}, ErrNotFound
	}
	if err != nil {
		return FileID{}, err
	}
	s, err := UnmarshalFileID(fileID.Sum)
	if err != nil {
		return FileID{}, err
	}
	return s, nil
}

// Delete deletes a file with a given ID. Returns jotfs.ErrNotFound if the file does not
// exist.
func (c *Client) Delete(file FileID) error {
	return c.DeleteWithContext(context.Background(), file)
}

// DeleteWithContext is the same as Delete with a user supplied context.
func (c *Client) DeleteWithContext(ctx context.Context, file FileID) error {
	_, err := c.iclient.Delete(ctx, &pb.FileID{Sum: file[:]})
	if e, ok := err.(twirp.Error); ok && e.Code() == twirp.NotFound {
		return ErrNotFound
	}
	return err
}

// getChunkerOptions gets the chunking options from the server.
func (c *Client) getChunkerOptions(ctx context.Context) (*fastcdc.Options, error) {
	params, err := c.iclient.GetChunkerParams(ctx, &pb.Empty{})
	if err != nil {
		return nil, err
	}
	return &fastcdc.Options{
		MinSize:       int(params.MinChunkSize),
		AverageSize:   int(params.AvgChunkSize),
		MaxSize:       int(params.MaxChunkSize),
		Normalization: int(params.Normalization),
	}, nil
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
