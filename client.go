package iotafs

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strconv"

	pb "github.com/iotafs/iotafs-go/internal/protos/upload"
	"github.com/iotafs/iotafs-go/internal/sum"

	"github.com/iotafs/fastcdc-go"
)

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

// Client implements methods used to interact with an IotaFS server.
type Client struct {
	host     string
	hclient  *http.Client
	iclient  pb.IotaFS
	cacheDir string
}

// New returns a new Client.
func New(host string) *Client {
	hclient := &http.Client{}
	return &Client{
		host:    host,
		hclient: hclient,
		iclient: pb.NewIotaFSProtobufClient(host, hclient),
	}
}

// UploadFile uploads a new file with a given name.
func (c *Client) UploadFile(r io.Reader, name string) error {
	ctx := context.Background()

	f, err := ioutil.TempFile("", "")
	if err != nil {
		return err
	}
	builder, err := newPackfileBuilder(f)
	if err != nil {
		return err
	}

	url := path.Join(c.host, "packfile")
	filenames := make(chan string)
	errc := make(chan error)
	go uploadPackfile(c.cacheDir, url, c.hclient, filenames, errc)

	chunker, err := fastcdc.NewChunker(r, defaultCDCOpts)
	if err != nil {
		return err
	}

	sums := make([][]byte, batchSize)
	eof := false
	i := 0
	for {
		chunk, err := chunker.Next()
		if err == io.EOF {
			eof = true
		}

		if i == batchSize || eof {
			i = 0
			exists, err := c.iclient.ChunksExist(ctx, &pb.ChunksExistRequest{Sums: sums})
			if err != nil {
				return err
			}
		}

		if eof {
			break
		}

		s := sum.Compute(chunk.Data)
		sums[i] = s[:]

		name := path.Join(c.cacheDir, s.AsHex())
		if err := ioutil.WriteFile(name, chunk.Data, 0644); err != nil {
			return err
		}

		i++
	}
}

type packer struct {
	maxPackfileSize uint64
	f               *os.File
	hclient         *http.Client
	builder         *packfileBuilder
	files           chan<- string
}

func (p *packer) addChunk(data []byte, s sum.Sum, mode compressMode) error {

	if p.builder.size()+uint64(len(data))+64 > p.maxPackfileSize {
		name := p.f.Name()
		if err := p.f.Close(); err != nil {
			return err
		}
		p.files <- name
	}

	if err := p.builder.append(data, s, mode); err != nil {
		return err
	}
}

func uploadPackfile(dir string, url string, client *http.Client, files <-chan string, errc chan<- error) {
	f := func(name string) error {
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
		info, err := f.Stat()
		if err != nil {
			return err
		}

		// Construct the request
		req, err := http.NewRequest("POST", url, f)
		if err != nil {
			return err
		}
		req.Header.Set("content-length", strconv.FormatInt(info.Size(), 10))
		req.Header.Set("x-iota-checksum", base64.StdEncoding.EncodeToString(s[:]))

		// Upload the file
		resp, err := client.Do(req)
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

		// We don't need the local packfile any longer
		if err := os.Remove(name); err != nil {
			return err
		}
		return nil
	}

	for name := range files {
		err := f(name)
		if err != nil {
			errc <- err
			return
		}
	}
}
