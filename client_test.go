package iotafs

import (
	"context"
	"os"
	"testing"
)

func TestUpload(t *testing.T) {

}

func BenchmarkUpload(b *testing.B) {
	client, err := New("http://localhost:6776")
	if err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	f, err := os.Open("random.txt")
	if err != nil {
		b.Fatal(err)
	}
	err = client.UploadWithContext(ctx, f, "random.txt", CompressNone)
	if err != nil {
		b.Fatal(err)
	}
}
