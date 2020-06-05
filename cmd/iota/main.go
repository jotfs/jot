package main

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jotfs/jot"
	"github.com/jotfs/jot/internal/admin"

	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

const (
	kiB = 1024
	miB = 1024 * kiB
	giB = 1024 * miB
)

type handler func(*jot.Client, *cli.Context) error

func getLatestVersion(client *jot.Client, name string) (jot.FileInfo, error) {
	it := client.Head(name, &jot.HeadOpts{Limit: 1})
	info, err := it.Next()
	if err == io.EOF {
		return jot.FileInfo{}, fmt.Errorf("file %s not found", name)
	}
	if err != nil {
		return jot.FileInfo{}, nil
	}
	return info, nil
}

func cp(client *jot.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() != 2 {
		return fmt.Errorf("two arguments expected")
	}

	src, srcRemote := isJotLocation(args.Get(0))
	dst, dstRemote := isJotLocation(args.Get(1))

	if srcRemote && dstRemote {
		// Copying from one JotFS location to another
		// TODO: allow user to specify version with --version flag
		// TODO: allow --all-versions flag
		latest, err := getLatestVersion(client, src)
		if err != nil {
			return err
		}
		fmt.Printf("copy: %s -> %s\n", src, dst)
		_, err = client.Copy(latest.FileID, dst)
		if err != nil {
			return err
		}

	} else if srcRemote && !dstRemote {
		// Copying from JotFS source to local destination (download)
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
		err = client.Download(latest.FileID, f)
		return mergeErrors(err, f.Close())

	} else if !srcRemote && dstRemote {
		// Copying from local source to JotFS destination (upload)
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
		return fmt.Errorf("at least one of <src> or <dst> must be an jot:// location")
	}

	return nil
}

func uploadFile(ctx context.Context, client *jot.Client, src string, dst string) error {
	f, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("unable to open %s: %v", src, err)
	}
	defer f.Close()
	fmt.Printf("upload: %s -> %s\n", src, dst)
	_, err = client.UploadWithContext(ctx, f, dst, jot.CompressNone)
	return err
}

// uploadRecursive recursively uploads the contents of a local directory.
func uploadRecursive(c *cli.Context, client *jot.Client, srcDir string, dstDir string) error {
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
func uploadDir(c *cli.Context, client *jot.Client, srcDir string, dstDir string) error {
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

func downloadFile(ctx context.Context, client *jot.Client, src jot.FileID, dst string) error {
	f, err := os.OpenFile(dst, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("unable to open %s: %v", src, err)
	}
	err = client.Download(src, f)
	return mergeErrors(err, f.Close())
}

// downloadRecursive recursively downloads the contents of a directory.
func downloadRecursive(c *cli.Context, client *jot.Client, prefix string, dstDir string) error {
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
		src jot.FileID
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
		case queue <- job{src: info.FileID, dst: dst}:
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

func ls(client *jot.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() > 1 {
		s := strings.Join(args.Slice(), ", ")
		return fmt.Errorf("expected 1 argument but received %d: %s", args.Len(), s)
	}
	var prefix string
	if args.Len() == 0 {
		prefix = "/"
	} else {
		prefix = args.Get(0)
	}

	format := "%-25s  %9s  %-8s  %s\n"
	fmt.Printf(format, "CREATED", "SIZE", "ID", "NAME")

	if c.Bool("recursive") {
		opts := &jot.ListOpts{Exclude: c.String("exclude"), Include: c.String("include")}
		it := client.List(prefix, opts)
		return lsOutput(it, format)
	}

	// Non-recursive ls
	// Get all files inside the directory
	exclude := path.Join(prefix, "*/*")
	opts := &jot.ListOpts{Exclude: exclude}
	it := client.List(prefix, opts)
	lsOutput(it, format)

	// Get all direct child directories
	dirsSeen := make(map[string]bool, 0)
	it = client.List(prefix, nil)
	prefix = cleanPath(prefix) // need to clean for TrimPrefix to work correctly below
	for {
		info, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		p := strings.TrimPrefix(info.Name, prefix)
		sp := strings.Split(p, "/")
		if len(sp) > 2 {
			_, seen := dirsSeen[sp[1]]
			if !seen {
				dirsSeen[sp[1]] = true
				fmt.Printf(format, "", "0 B  ", "DIR", path.Join(prefix, sp[1])+"/")
			}
		}
	}
	return nil
}

// lsOutput prints the output from a FileIterator
func lsOutput(it jot.FileIterator, format string) error {
	for {
		info, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		ts := info.CreatedAt.Local().Format(time.RFC3339)
		s := hex.EncodeToString(info.FileID.Marshal())
		fmt.Printf(format, ts, humanBytes(info.Size), s[:8], info.Name)
	}
	return nil
}

// cleanPath adds a leading slash if name doesn't already have one, and removes a
// trailing if present, and all leading and trailing whitespace.
func cleanPath(name string) string {
	if name == "" {
		return name
	}
	name = strings.TrimSpace(name)
	if !strings.HasPrefix(name, "/") {
		name = "/" + name
	}
	if strings.HasSuffix(name, "/") {
		name = strings.TrimSuffix(name, "/")
	}
	return name
}

func rm(client *jot.Client, c *cli.Context) error {
	args := c.Args()

	// TODO: implement --version flag (only one arg allowed in this case)
	// TODO: handle trailing slash

	for _, name := range args.Slice() {
		var it jot.FileIterator
		if c.Bool("recursive") {
			opts := &jot.ListOpts{Exclude: c.String("exclude"), Include: c.String("include")}
			it = client.List(name, opts)
		} else if c.Bool("all-versions") {
			it = client.Head(name, nil)
		} else {
			// Just the latest version
			it = client.Head(name, &jot.HeadOpts{Limit: 1})
		}

		for {
			info, err := it.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}
			s := hex.EncodeToString(info.FileID.Marshal())
			fmt.Printf("delete: %s %s\n", info.Name, s[:8])
			if err := client.Delete(info.FileID); err != nil {
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
		return fmt.Sprintf("%5.1d B", size)
	}
	if size < miB {
		return fmt.Sprintf("%5.1f KiB", float64(size)/kiB)
	}
	if size < giB {
		return fmt.Sprintf("%5.1f MiB", float64(size)/miB)
	}
	return fmt.Sprintf("%5.1f GiB", float64(size)/giB)
}

func isJotLocation(s string) (string, bool) {
	if strings.HasPrefix(s, "jot://") {
		return s[6:], true
	}
	return s, false
}

func main() {
	var client *jot.Client
	var adminC *admin.Client
	var endpoint string

	app := cli.NewApp()
	app.Name = "jot"
	app.Usage = "A CLI tool for working with an JotFS server"
	app.Description = description
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
		&cli.StringFlag{
			Name:  "endpoint",
			Usage: "set the server endpoint",
		},
	}

	// Get the server endpoint, either from the command line option or config file, and
	// initialize the client
	app.Before = func(c *cli.Context) error {
		if isOneOf(c.Args().First(), []string{"help", "h"}) {
			// Return early if it's the help subcommand
			return nil
		}
		var err error
		endpoint, err = func() (string, error) {
			endpoint := c.String("endpoint")
			if endpoint != "" {
				return endpoint, nil
			}
			cfgName := c.String("config")
			if cfgName == "" {
				cfgName = getConfigFile()
			}
			if cfgName == "" {
				return "", errors.New("unable to find config file")
			}
			profileName := c.String("profile")
			profile, err := loadConfig(cfgName, profileName)
			if err != nil {
				return "", err
			}
			return profile.Endpoint, nil
		}()
		if err != nil {
			return err
		}
		client, err = jot.New(endpoint, nil)
		if err != nil {
			return err
		}
		adminC, err = admin.New(endpoint)
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
			Name:        "cp",
			Usage:       "copy files to / from JotFS",
			UsageText:   "jot cp <src> <dst>",
			Description: cpDescription,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r"},
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
					Usage: "max. number of concurrent operations",
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
			UsageText: "jot ls <prefix>",
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
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r"},
					Usage:   "list recursively",
				},
			},
		},
		{
			Name:        "rm",
			Usage:       "remove files",
			UsageText:   "jot rm <file>...",
			Description: rmDescription,
			Flags: []cli.Flag{
				&cli.BoolFlag{
					Name:  "all-versions",
					Usage: "remove all versions",
					Value: false,
				},
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r"},
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
		{
			Name:  "admin",
			Usage: "server administration",
			Subcommands: []*cli.Command{
				{
					Name:  "start-vacuum",
					Usage: "manually start the server vacuum process",
					Action: func(c *cli.Context) error {
						id, err := adminC.StartVacuum(c.Context)
						if err == nil {
							fmt.Printf("vacuum %s started\n", id)
						}
						if errors.Is(err, admin.ErrVacuumInProgress) {
							fmt.Println("vacuum already in progress")
							return nil
						}
						return err
					},
				},
				{
					Name:  "vacuum-status",
					Usage: "get the status of a vacuum",
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "id",
							Usage:    "vacuum ID",
							Required: true,
						},
					},
					Action: func(c *cli.Context) error {
						vacuum, err := adminC.VacuumStatus(c.Context, c.String("id"))
						if err != nil {
							return err
						}
						if vacuum.Status == "RUNNING" {
							fmt.Println(vacuum.Status)
							return nil
						}
						elapsed := vacuum.CompletedAt.Sub(vacuum.StartedAt)
						fmt.Printf("%s (%.1f seconds)\n", vacuum.Status, elapsed.Seconds())
						return nil
					},
				},
				{
					Name:  "stats",
					Usage: "view server summary statistics",
					Action: func(c *cli.Context) error {
						stats, err := adminC.ServerStats(c.Context)
						if err != nil {
							return err
						}
						compRatio := float64(stats.TotalFilesSize) / float64(stats.TotalDataSize)
						tFilesSize := strings.TrimSpace(humanBytes(stats.TotalFilesSize))
						tDataSize := strings.TrimSpace(humanBytes(stats.TotalDataSize))
						fmt.Printf("%-23s = %d\n", "Number of files", stats.NumFiles)
						fmt.Printf("%-23s = %d\n", "Number of file versions", stats.NumFileVersions)
						fmt.Printf("%-23s = %s\n", "Total files size", tFilesSize)
						fmt.Printf("%-23s = %s\n", "Total data size", tDataSize)
						fmt.Printf("%-23s = %.1f\n", "Compression ratio", compRatio)
						return nil
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

var description = `
   jot will look for its configuration file at $HOME/.jot/config.toml 
   by default. Alternatively, its path may be specified by setting the 
   JOT_CONFIG_FILE environment variable, or with the --config option.
   The server endpoint URL may be overridden with the --endpoint option.`

var cpDescription = `
   At least one of <src> or <dst> must be prefixed with jot:// to
   signify the operation as an upload, download or copy.

EXAMPLES:
   
   Download a single file:

      jot cp jot://images/bird.png ./the_bird.png

   Download recursive (local directories will be created):

      jot cp -r jot://images images
	  
   Upload a single file:

      jot cp test.csv jot://data/test.csv  

   Upload recursive:

      jot cp -r datasets/ jot://datasets

   Exclude certain files:

      jot cp -r --exclude "/img/*" --include "/img/a.png" jot:// files/ 

   Copy a file from one JotFS location to another:

      jot cp jot://test.csv jot://data/test-copy.csv
`

var rmDescription = `
   If versioning is enabled, only the latest version of each matching
   file will be removed. Set the --all-versions flag to remove all 
   versions. To recursively remove files, set the --recursive flag.

   WARNING: --recursive will always delete all versions of each file.

EXAMPLES:

   Remove files:
	  
      jot rm bird.png img/dog.png

   Remove all files in img/ except those in img/birds:

      jot rm -r --include "img/*" --exclude "img/birds/*" img/
`

var lsDescription = `

`
