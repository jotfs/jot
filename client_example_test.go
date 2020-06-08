package jot_test

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/jotfs/jot"
)

func ExampleNew() {
	jot.New("http://example.com:6777", http.DefaultClient, nil)
}

func ExampleNew_options() {
	jot.New("https://jotfs.example.com", http.DefaultClient, &jot.Options{
		Compression: jot.CompressNone,
	})
}

func ExampleClient_Upload() {
	client, err := jot.New("http://localhost:6777", http.DefaultClient, nil)
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
	client, err := jot.New("http://localhost:6777", http.DefaultClient, nil)
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
