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
	"time"

	"github.com/google/uuid"
	pb "github.com/iotafs/iotafs-go/internal/protos/upload"
	"github.com/iotafs/iotafs-go/internal/sum"

	"github.com/iotafs/fastcdc-go"
)

var ErrNotFound = errors.New("not found")

const (
	kiB             = 1024
	miB             = 1024 * kiB
	maxPackfileSize = 128 * miB
	batchSize       = 8
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

// New returns a new Client.
func New(host string) (*Client, error) {
	url, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	hclient := &http.Client{}
	return &Client{
		host:    *url,
		hclient: hclient,
		iclient: pb.NewIotaFSProtobufClient(host, hclient),
	}, nil
}

// UploadWithContext uploads a new file with a given name.
func (c *Client) UploadWithContext(ctx context.Context, r io.Reader, dst string, mode compressMode) error {
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
	dir := path.Join(c.cacheDir, id.String())
	if err := os.Mkdir(dir, 0744); err != nil {
		return err
	}
	defer os.RemoveAll(dir) // TODO: log error

	packer, err := newPacker(dir, packFiles)

	chunker, err := fastcdc.NewChunker(r, defaultCDCOpts)
	if err != nil {
		return err
	}

	sums := make([][]byte, 0)

	// Upload all new chunks in the file as packfiles to the server
	err = func() error {
		defer close(packFiles) // ensure the uploader goroutine terminates
		batch := make([]sum.Sum, batchSize)
		eof := false
		for i := 0; ; i++ {
			chunk, err := chunker.Next()
			if err == io.EOF {
				eof = true
				batch = batch[:i]
			} else if err != nil {
				return err
			}

			if i == batchSize || eof {
				// Check which chunks in the batch need to be added to a packfile
				resp, err := c.iclient.ChunksExist(ctx, &pb.ChunksExistRequest{Sums: sumsToBytes(batch)})
				if err != nil {
					return err
				}
				for j, sum := range batch {
					data, err := popChunk(dir, sum)
					if err != nil {
						return err
					}
					if resp.Exists[j] {
						continue
					}
					if err := packer.addChunk(data, sum, mode); err != nil {
						return err
					}
				}
				i = 0 // start a new batch
			}
			if eof {
				break
			}

			s := sum.Compute(chunk.Data)
			batch[i] = s
			sums = append(sums, s[:])

			if err := saveChunk(dir, chunk.Data, s); err != nil {
				return err
			}
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
	_, err = c.iclient.CreateFile(ctx, &pb.File{Name: dst, Sums: sums})
	if err != nil {
		return err
	}

	return nil
}

// uploadPackfiles listens for packfiles on a channel and uploads them. It signals
// completion (a nil error) or failure on the done channel.
func (c *Client) uploadPackfiles(ctx context.Context, files <-chan string, done chan<- error) {
	u := c.host
	u.Path = path.Join(u.Path, "packfile")
	url := u.String()

	upload := func(name string) error {
		// Get the packfile checksum from its file name
		s, err := sum.FromHex(path.Base(name))
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
		os.Remove(name) // get rid of the local packfile TODO: log error
		if err != nil {
			done <- err
		}
	}
	done <- nil
}

// saveChunk writes a chunk of data to the cache directory.
func saveChunk(dir string, data []byte, s sum.Sum) error {
	name := path.Join(dir, s.AsHex())
	return ioutil.WriteFile(name, data, 0644)
}

// popChunk removes a chunk from the cache directory and returns it.
func popChunk(dir string, s sum.Sum) ([]byte, error) {
	name := path.Join(dir, s.AsHex())
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return data, os.Remove(name)
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
	name := path.Join(p.dir, id.String())
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
	packName := path.Join(p.dir, packSum.AsHex())
	if err := os.Rename(p.f.Name(), packName); err != nil {
		return err
	}
	p.packFiles <- packName

	return nil
}

// addChunk adds a chunk to a packfile owned by the packer.
func (p *packer) addChunk(data []byte, sum sum.Sum, mode compressMode) error {
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

func (c *Client) ListFiles(prefix string) ([]FileInfo, error) {
	ctx := context.Background()
	files, err := c.iclient.ListFiles(ctx, &pb.Prefix{Prefix: prefix})
	if err != nil {
		return nil, err
	}

	infos := make([]FileInfo, len(files.Infos))
	for i, info := range files.Infos {
		fmt.Printf("%x\n", info.Sum)
		s, err := sum.FromBytes(info.Sum)
		if err != nil {
			return nil, err
		}
		infos[i] = FileInfo{
			Name:      info.Name,
			CreatedAt: time.Unix(0, info.CreatedAt),
			Size:      info.Size,
			Sum:       s,
		}
	}

	return infos, nil
}

func (c *Client) HeadFile(name string) ([]FileInfo, error) {
	// TODO: implement pagination

	ctx := context.Background()
	files, err := c.iclient.HeadFile(ctx, &pb.HeadFileRequest{Name: name, Limit: 1000})
	if err != nil {
		return nil, err
	}

	infos := make([]FileInfo, len(files.Info))
	for i, info := range files.Info {
		s, err := sum.FromBytes(info.Sum)
		if err != nil {
			return nil, err
		}
		infos[i] = FileInfo{
			Name:      info.Name,
			CreatedAt: time.Unix(0, info.CreatedAt),
			Size:      info.Size,
			Sum:       s,
		}
	}

	return infos, nil
}

type FileInfo struct {
	Name      string
	CreatedAt time.Time
	Size      uint64
	Sum       sum.Sum
}

func (c *Client) Download(fileID sum.Sum, dst string) error {
	ctx := context.Background()
	resp, err := c.iclient.Download(ctx, &pb.FileID{Sum: fileID[:]})
	if err != nil {
		return err
	}

	f, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", dst, err)
	}

	// Download the data for each packfile section using the provided URLs, and use
	// the data to construct the original file.
	err = func() error {
		for i, s := range resp.Sections {
			err := c.downloadSection(f, s)
			if err != nil {
				return fmt.Errorf("section %d: %w", i, err)
			}
		}
		return nil
	}()
	cerr := f.Close()
	if err != nil || cerr != nil {
		return mergeErrors(err, cerr)
	}

	return nil
}

func (c *Client) downloadSection(f *os.File, s *pb.Section) error {
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
		w := io.MultiWriter(f, h)
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

func (c *Client) Copy(src sum.Sum, dst string) (sum.Sum, error) {
	ctx := context.Background()
	fileID, err := c.iclient.Copy(ctx, &pb.CopyRequest{SrcId: src[:], Dst: dst})
	// TODO: return NotFound if twirp.NotFoundError
	if err != nil {
		return sum.Sum{}, err
	}
	s, err := sum.FromBytes(fileID.Sum)
	if err != nil {
		return sum.Sum{}, err
	}
	return s, nil
}

func (c *Client) Delete(fileID sum.Sum) error {
	ctx := context.Background()
	// TODO: return NotFound if twirp.NotFoundError
	_, err := c.iclient.Delete(ctx, &pb.FileID{Sum: fileID[:]})
	return err
}
