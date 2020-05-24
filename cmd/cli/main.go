package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/iotafs/iotafs-go"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

const (
	kiB = 1024
	miB = 1024 * kiB
	giB = 1024 * miB
)

type handler func(*iotafs.Client, *cli.Context) error

func getLatestVersion(client *iotafs.Client, name string) (iotafs.FileInfo, error) {
	versions, err := client.Head(name, 1)
	if err != nil {
		return iotafs.FileInfo{}, err
	}
	if len(versions) == 0 {
		return iotafs.FileInfo{}, fmt.Errorf("file %s not found", name)
	}
	return versions[0], nil
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
		if err := client.Download(latest.Sum, dstEx); err != nil {
			return err
		}
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
				return uploadRecursive(c.Context, client, srcEx, dst)
			}
			return uploadDir(c.Context, client, srcEx, dst, iotafs.CompressNone)
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
func uploadRecursive(ctx context.Context, client *iotafs.Client, srcDir string, dstDir string) error {
	if strings.HasSuffix(srcDir, "..") {
		return errors.New("src cannot end in \"..\"")
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
		return uploadFile(ctx, client, src, dst)
	})
	return err
}

// uploadDir uploads the contents of a local directory.
func uploadDir(ctx context.Context, client *iotafs.Client, srcDir string, dstDir string, mode iotafs.CompressMode) error {
	entries, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		src := filepath.Join(srcDir, entry.Name())
		dst := filepath.Join(dstDir, entry.Name())
		return uploadFile(ctx, client, src, dst)
	}
	return nil
}

func ls(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() != 1 {
		return fmt.Errorf("only 1 argument expected")
	}
	pattern := args.Get(0)

	res, err := client.List(pattern)
	if err != nil {
		return err
	}
	format := "%-25s  %9s  %-8s  %s\n"
	fmt.Printf(format, "CREATED", "SIZE", "ID", "NAME")
	for _, row := range res {
		ts := row.CreatedAt.Local().Format(time.RFC3339)
		s := row.Sum.AsHex()[:8]
		fmt.Printf(format, ts, humanBytes(row.Size), s, row.Name)
	}

	return nil
}

func rm(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()

	// TODO: implement --version flag (only one arg allowed in this case)

	for _, name := range args.Slice() {
		limit := uint64(1)
		if c.Bool("all-versions") {
			limit = 1000 // TODO: pagination
		}
		versions, err := client.Head(name, limit)
		if err != nil {
			return err
		}
		for i := range versions {
			v := versions[i]
			fmt.Printf("delete: %s %s\n", name, v.Sum.AsHex()[:8])
			if err := client.Delete(v.Sum); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeErrorf(format string, args ...interface{}) {
	_, err := os.Stderr.WriteString(fmt.Sprintf(format, args...))
	if err != nil {
		panic(err)
	}
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

func toSentence(s string) string {
	if s == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(s)
	if r != utf8.RuneError {
		s = string(unicode.ToUpper(r)) + s[size:]
	}
	r, _ = utf8.DecodeLastRuneInString(s)
	if r != utf8.RuneError && r != rune('.') {
		s = s + "."
	}
	return s
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

func main() {

	client, err := iotafs.New("http://localhost:6776")
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
		return
	}

	app := cli.NewApp()
	app.Name = "IotaFS Client"
	app.ExitErrHandler = func(c *cli.Context, err error) {
		os.Stderr.WriteString("Error: " + toSentence(err.Error()) + "\n")
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
				&cli.StringFlag{
					Name:  "compression",
					Value: "zstd",
				},
				&cli.BoolFlag{
					Name:    "recursive",
					Aliases: []string{"r", "R"},
					Usage:   "copy directory recursively",
				},
			},
			Action: makeAction(cp),
		},
		{
			Name:      "ls",
			Usage:     "list files",
			UsageText: "iota ls <pattern>",
			Action:    makeAction(ls),
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
			},
			Action: makeAction(rm),
		},
	}

	app.Run(os.Args)
}
