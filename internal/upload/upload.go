// package upload

// import (
// 	"bytes"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"net/http"
// 	"net/url"
// 	"path"

// 	"github.com/iotafs/fastcdc-go"
// 	pb "github.com/iotafs/iotafs-go/internal/protos/upload"
// 	"github.com/iotafs/iotafs-go/internal/sum"
// )

// const (
// 	kiB             = 1024
// 	miB             = 1024 * kiB
// 	maxPackfileSize = 128 * miB
// )

// var defaultCDCOpts = fastcdc.Options{
// 	MinSize:    256 * kiB,
// 	MaxSize:    8 * miB,
// 	NormalSize: 1 * miB,
// 	SmallBits:  22,
// 	LargeBits:  18,
// }

// type Config struct {
// 	Host string
// }

// type Client struct {
// 	cfg    Config
// 	client *http.Client
// }

// func (c *Client) Upload(name string, r io.Reader) error {
// 	pclient := pb.NewIotaFSProtobufClient(c.cfg.Host, c.client)

// 	chunker, err := fastcdc.NewChunker(r, defaultCDCOpts)
// 	if err != nil {
// 		return err
// 	}

// 	f, err := ioutil.TempFile("", "iotafs-pack-")
// 	if err != nil {
// 		return err
// 	}

// 	queue := make(chan []byte)
// 	url := chunkUploadURL(c.cfg.Host, uploadID.Id)
// 	uploader := newUploader(url, c.client, queue)
// 	go uploader.upload()

// 	err = func() error {
// 		// Lag the chunks by one so we know when the final chunk is reached
// 		chunk, err := chunker.Next()
// 		if err == io.EOF {
// 			return fmt.Errorf("reader is empty")
// 		} else if err != nil {
// 			return err
// 		}
// 		buf := make([]byte, 0, len(chunk.Data))
// 		copy(buf, chunk.Data)

// 		var final bool
// 		for i := uint64(0); ; i++ {
// 			nextChunk, err := chunker.Next()
// 			if err == io.EOF {
// 				final = true
// 			} else if err != nil {
// 				return err
// 			}

// 			s := sum.Compute(chunk.Data)
// 			addChunk, err := pclient.AddChunk(ctx, &pb.Chunk{
// 				UploadId: uploadID.Id,
// 				Sequence: i,
// 				Size:     uint64(chunk.Length),
// 				Sum:      s[:],
// 				Final:    final,
// 			})
// 			if err != nil {
// 				return err
// 			}

// 			// TODO: chunk upload here can be done asynchronously on a work queue
// 			if addChunk.Upload {
// 				buf := make([]byte, chunk.Length)
// 				copy(buf, chunk.Data)
// 			}

// 			chunk = nextChunk
// 			data = data[:0]
// 			data = append(data, chunk.Data...)
// 			if final {
// 				return nil
// 			}
// 		}
// 	}()
// 	if err != nil {
// 		err = fmt.Errorf("aborting upload: %v", err)
// 		_, aerr := pclient.Abort(ctx, uploadID)
// 		if aerr != nil {
// 			err = fmt.Errorf("%v. Also abort error: %v", err, aerr)
// 		}
// 		return err
// 	}

// 	_, err = pclient.Complete(ctx, &pb.File{UploadId: uploadID.Id, Name: name})
// 	// TODO: catch twirp.DeadlineExceeded and retry
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// type uploader struct {
// 	url    string
// 	queue  <-chan []byte
// 	done   chan error
// 	client *http.Client
// }

// func newUploader(url string, c *http.Client, queue <-chan []byte) *uploader {
// 	done := make(chan error, 1)
// 	return &uploader{url, queue, done, c}
// }

// func (u *uploader) upload() {
// 	for data := range u.queue {
// 		r := bytes.NewReader(data)
// 		resp, err := u.client.Post(u.url, "application/octet-stream", r)
// 		if err != nil {
// 			u.done <- err
// 			return
// 		}
// 		defer resp.Body.Close()
// 		if resp.StatusCode != http.StatusAccepted {
// 			b, _ := ioutil.ReadAll(resp.Body)
// 			msg := string(b)
// 			err = fmt.Errorf("chunk upload failed (%d): %s", resp.StatusCode, msg)
// 			u.done <- err
// 			return
// 		}
// 	}
// 	u.done <- nil
// }

// func (u *uploader) isDone() <-chan error {
// 	return u.done
// }

// func chunkUploadURL(host, id string) string {
// 	s := path.Join(host, "chunk")
// 	return s + "?upload_id=" + url.QueryEscape(id)
// }
