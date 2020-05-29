package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/iotafs/iotafs-go"
	"github.com/iotafs/iotafs-go/internal/sum"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	kiB = 1024
	miB = 1024 * kiB
	giB = 1024 * miB
)

type handler func(*iotafs.Client, *cli.Context) error

func getLatestVersion(client *iotafs.Client, name string) (iotafs.FileInfo, error) {
	it := client.Head(name, &iotafs.IteratorOpts{Limit: 1})
	info, err := it.Next()
	if err == io.EOF {
		return iotafs.FileInfo{}, fmt.Errorf("file %s not found", name)
	}
	if err != nil {
		return iotafs.FileInfo{}, nil
	}
	return info, nil
}

func cp(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() != 2 {
		return fmt.Errorf("two arguments expected")
	}

	src, srcRemote := isIotaLocation(args.Get(0))
	dst, dstRemote := isIotaLocation(args.Get(1))

	if srcRemote && dstRemote {
		// Copying from one IotaFS location to another
		// TODO: allow user to specify version with --version flag
		// TODO: allow --all-versions flag
		latest, err := getLatestVersion(client, src)
		if err != nil {
			return err
		}
		fmt.Printf("copy: %s -> %s\n", src, dst)
		_, err = client.Copy(latest.Sum, dst)
		if err != nil {
			return err
		}

	} else if srcRemote && !dstRemote {
		// Copying from IotaFS source to local destination (download)
		// TODO: allow user to specify version with --version flag
		dstEx, err := homedir.Expand(dst)
		if err != nil {
			return fmt.Errorf("invalid path %s", dst)
		}

		if c.Bool("recursive") {
			return downloadRecursive(c, client, src, dst)
		}

		// Check that the directory specified by dst exists
		info, err := os.Stat(dstEx)
		if os.IsNotExist(err) {
			dir := filepath.Dir(dstEx)
			info, err := os.Stat(dir)
			if os.IsNotExist(err) {
				return fmt.Errorf("invalid path %s: directory does not exist", dst)
			}
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return fmt.Errorf("invalid path %s: directory does not exist", dir)
			}
		} else if err != nil {
			return err
		} else {
			if info.IsDir() {
				// The user has not explicitly specified a filename. Set it to the
				// base name of the remote file
				dstEx = filepath.Join(dstEx, filepath.Base(src))
			}
			// dstEx exists and is a file. It will be overwritten
		}

		latest, err := getLatestVersion(client, src)
		if err != nil {
			return err
		}

		fmt.Printf("download: %s -> %s\n", src, dstEx)
		f, err := os.OpenFile(dstEx, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		err = client.Download(latest.Sum, f)
		return mergeErrors(err, f.Close())

	} else if !srcRemote && dstRemote {
		// Copying from local source to Iota destination (upload)
		src := filepath.Clean(src)
		srcEx, err := homedir.Expand(src)
		if err != nil {
			return fmt.Errorf("invalid path %q", src)
		}

		// Check that the file / directory exists
		info, err := os.Stat(srcEx)
		if err != nil {
			return fmt.Errorf("invalid path %q", src)
		}

		if info.IsDir() {
			if c.Bool("recursive") {
				return uploadRecursive(c, client, srcEx, dst)
			}
			return uploadDir(c, client, srcEx, dst)
		}

		return uploadFile(c.Context, client, srcEx, dst)

	} else {
		return fmt.Errorf("at least one of <src> or <dst> must be an iota:// location")
	}

	return nil
}

func uploadFile(ctx context.Context, client *iotafs.Client, src string, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open %s: %v", src, err)
	}
	defer f.Close()
	fmt.Printf("upload: %s -> %s\n", src, dst)
	return client.UploadWithContext(ctx, f, dst, iotafs.CompressNone)
}

// uploadRecursive recursively uploads the contents of a local directory.
func uploadRecursive(c *cli.Context, client *iotafs.Client, srcDir string, dstDir string) error {
	if strings.HasSuffix(srcDir, "..") {
		return errors.New("src cannot end in \"..\"")
	}

	type job struct {
		src string
		dst string
	}
	queue := make(chan job)

	g, ctx := errgroup.WithContext(c.Context)
	numWorkers := c.Uint("concurrency")
	if numWorkers < 0 {
		return errors.New("option --concurrency must be at least 1")
	}
	for i := uint(0); i < numWorkers; i++ {
		g.Go(func() error {
			for job := range queue {
				err := uploadFile(c.Context, client, job.src, job.dst)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	err := filepath.Walk(srcDir, func(src string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(srcDir, src)
		if err != nil {
			return err
		}
		dst := filepath.Join(dstDir, rel)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case queue <- job{src, dst}:
			return nil
		}
	})

	close(queue)

	return mergeErrors(g.Wait(), err)
}

// uploadDir uploads the contents of a local directory.
func uploadDir(c *cli.Context, client *iotafs.Client, srcDir string, dstDir string) error {
	entries, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}

	type job struct {
		src string
		dst string
	}
	queue := make(chan job)

	g, ctx := errgroup.WithContext(c.Context)
	numWorkers := c.Int("concurrency")
	if numWorkers <= 0 {
		return errors.New("option --concurrency must be at least 1")
	}
	for i := 0; i < numWorkers; i++ {
		g.Go(func() error {
			for job := range queue {
				err := uploadFile(c.Context, client, job.src, job.dst)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(srcDir, entry.Name())
		dst := filepath.Join(dstDir, entry.Name())
		var err error
		select {
		case <-ctx.Done():
			err = ctx.Err()
		case queue <- job{src, dst}:
			err = nil
		}
		if err != nil {
			break
		}
	}

	close(queue)

	return g.Wait()
}

func downloadFile(ctx context.Context, client *iotafs.Client, src sum.Sum, dst string) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s: %v", src, err)
	}
	err = client.Download(src, f)
	return mergeErrors(err, f.Close())
}

// downloadRecursive recursively downloads the contents of a directory.
func downloadRecursive(c *cli.Context, client *iotafs.Client, prefix string, dstDir string) error {
	// Create the source directory if it doesn't exist
	exists, err := dirExists(dstDir)
	if err != nil {
		return err
	}
	if !exists {
		if err := os.MkdirAll(dstDir, 0744); err != nil {
			return fmt.Errorf("making directory: %v", err)
		}
	}

	numWorkers := c.Int("concurrency")
	if numWorkers <= 0 {
		return errors.New("option --concurrency must be at least 1")
	}

	// job stores the checksum of a file to download and the local dst to save it to
	type job struct {
		src sum.Sum
		dst string
	}
	queue := make(chan job)

	// Launch numWorkers goroutines which execute jobs taken from the work queue
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()
	g, gctx := errgroup.WithContext(ctx)
	for i := 0; i < numWorkers; i++ {
		g.Go(func() error {
			for job := range queue {
				// create the directory for the file if it doesn't exist
				dir := filepath.Dir(job.dst)
				exists, err := dirExists(dir)
				if err != nil {
					return err
				}
				if !exists {
					if err = os.MkdirAll(dir, 0744); err != nil {
						return fmt.Errorf("making directory %s: %v", dir, err)
					}
				}

				err = downloadFile(gctx, client, job.src, job.dst)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}

	// Add each file matching the prefix to the download queue
	it := client.List(prefix, nil)
	for {
		info, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Cancel any downloads currently running
			cancel()
			return err
		}

		dst := filepath.Join(dstDir, filepath.FromSlash(info.Name))
		select {
		case <-gctx.Done():
			break
		case queue <- job{src: info.Sum, dst: dst}:
			fmt.Printf("download: %s -> %s\n", info.Name, dst)
		}
	}

	close(queue)

	return g.Wait()
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

func ls(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() != 1 {
		return fmt.Errorf("only 1 argument expected")
	}
	pattern := args.Get(0)

	it := client.ListFilter(pattern, c.String("exclude"), c.String("include"), nil)
	format := "%-25s  %9s  %-8s  %s\n"
	fmt.Printf(format, "CREATED", "SIZE", "ID", "NAME")
	for {
		info, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		ts := info.CreatedAt.Local().Format(time.RFC3339)
		s := info.Sum.AsHex()[:8]
		fmt.Printf(format, ts, humanBytes(info.Size), s, info.Name)
	}

	return nil
}

func rm(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()

	// TODO: implement --version flag (only one arg allowed in this case)
	// TODO: handle trailing slash

	for _, name := range args.Slice() {
		var it iotafs.FileIterator
		if c.Bool("recursive") {
			it = client.ListFilter(name, c.String("exclude"), c.String("include"), nil)
		} else if c.Bool("all-versions") {
			it = client.Head(name, nil)
		} else {
			// Just the latest version
			it = client.Head(name, &iotafs.IteratorOpts{Limit: 1})
		}

		for {
			info, err := it.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			fmt.Printf("delete: %s %s\n", info.Name, info.Sum.AsHex()[:8])
			if err := client.Delete(info.Sum); err != nil {
				return err
			}
		}
	}

	return nil
}

func dirExists(name string) (bool, error) {
	info, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}

func humanBytes(size uint64) string {
	if size < kiB {
		return fmt.Sprintf("%5.1d B  ", size)
	}
	if size < miB {
		return fmt.Sprintf("%5.1f KiB", float64(size)/kiB)
	}
	if size < giB {
		return fmt.Sprintf("%5.1f MiB", float64(size)/miB)
	}
	return fmt.Sprintf("%5.1f GiB", float64(size)/giB)
}

func isIotaLocation(s string) (string, bool) {
	if strings.HasPrefix(s, "iota://") {
		return s[6:], true
	}
	return s, false
}

var client *iotafs.Client

func main() {

	app := cli.NewApp()
	app.Name = "IotaFS Client"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "config",
			Usage: "path to config file",
		},
		&cli.StringFlag{
			Name:  "profile",
			Usage: "config profile to use",
			Value: "default",
		},
	}

	app.Before = func(c *cli.Context) error {
		// Load the config and initialize the client
		if isOneOf(c.Args().First(), []string{"help", "h"}) {
			// Return early if it's the help subcommand
			return nil
		}
		cfgName := c.String("config")
		if cfgName == "" {
			cfgName = getConfigFile()
		}
		if cfgName == "" {
			return errors.New("unable to find config file")
		}

		profileName := c.String("profile")
		p, err := loadConfig(cfgName, profileName)
		if err != nil {
			return err
		}

		client, err = iotafs.New(p.Endpoint)
		return err
	}

	app.ExitErrHandler = func(c *cli.Context, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	makeAction := func(h handler) cli.ActionFunc {
		return func(c *cli.Context) error { return h(client, c) }
	}

	app.Commands = []*cli.Command{
		{
			Name:      "cp",
			Usage:     "copy files to / from IotaFS",
			UsageText: "iota cp <src> <dst>",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r", "R"},
					Usage:   "copy directory recursively",
				},
				&cli.StringFlag{
					Name:  "exclude",
					Usage: "exclude files matching the given pattern",
				},
				&cli.StringFlag{
					Name:  "include",
					Usage: "don't exclude files matching the given pattern",
				},
				&cli.UintFlag{
					Name:  "concurrency",
					Value: 3,
				},
				&cli.StringFlag{
					Name:  "compression",
					Value: "zstd",
				},
			},
			Action: makeAction(cp),
		},
		{
			Name:      "ls",
			Usage:     "list files",
			UsageText: "iota ls <pattern>",
			Action:    makeAction(ls),
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "exclude",
					Usage: "exclude files matching the given pattern",
				},
				&cli.StringFlag{
					Name:  "include",
					Usage: "don't exclude files matching the given pattern",
				},
			},
		},
		{
			Name:      "rm",
			Usage:     "remove files",
			UsageText: "iota rm <file>...",
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "all-versions",
					Usage: "remove all versions",
					Value: false,
				},
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r", "R"},
					Usage:   "remove recursively",
				},
				&cli.StringFlag{
					Name:  "exclude",
					Usage: "exclude files matching the given pattern",
				},
				&cli.StringFlag{
					Name:  "include",
					Usage: "don't exclude files matching the given pattern",
				},
			},
			Action: makeAction(rm),
		},
	}

	app.Run(os.Args)
}
