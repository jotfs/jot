package iotafs_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/iotafs/iotafs-go"
)

func ExampleNew() {
	iotafs.New("http://example.com:6777", nil)
}

func ExampleNew_options() {
	iotafs.New("https://iotafs.example.com", &iotafs.Options{
		Compression: iotafs.CompressNone,
	})
}

func ExampleClient_Upload() {
	client, err := iotafs.New("http://localhost:6777", nil)
	if err != nil {
		log.Fatal(err)
	}

	r := strings.NewReader("Hello World!")
	fileID, err := client.Upload(r, "/notes/today.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("fileID = %x\n", fileID.Marshal())
}

func ExampleClient_Download() {
	client, err := iotafs.New("http://localhost:6777", nil)
	if err != nil {
		log.Fatal(err)
	}

	r := strings.NewReader("Hello World!")
	fileID, err := client.Upload(r, "/notes/today.txt")
	if err != nil {
		log.Fatal(err)
	}

	var w bytes.Buffer
	err = client.Download(fileID, &w)
	if err != nil {
		log.Fatal(err)
	}
}
