package main

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/iotafs/iotafs-go"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
)

type handler func(*iotafs.Client, *cli.Context) error

func cp(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() != 2 {
		return fmt.Errorf("two arguments expected")
	}

	src, srcRemote := isIotaLocation(args.Get(0))
	dst, dstRemote := isIotaLocation(args.Get(1))

	if srcRemote && dstRemote {
		// Copying from one Iota location to another
		// TODO: allow user to specify version with --version flag
		versions, err := client.HeadFile(src)
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			return fmt.Errorf("file %s not found on remote", src)
		}
		latest := versions[0]

		newID, err := client.Copy(latest.Sum, dst)
		if err != nil {
			return err
		}
		fmt.Printf("%s copied to %s with version ID %s\n", src, dst, newID.AsHex())

	} else if srcRemote && !dstRemote {
		// Copying from Iota source to local destination (download)
		dstEx, err := homedir.Expand(dst)
		if err != nil {
			return fmt.Errorf("invalid path %q", dst)
		}

		versions, err := client.HeadFile(src)
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			return fmt.Errorf("file %s not found on remote", src)
		}
		latest := versions[0]

		// TODO: check destination directory exists
		// TODO: allow user to specify version with --version flag

		if err := client.Download(latest.Sum, dstEx); err != nil {
			return err
		}
	} else if !srcRemote && dstRemote {
		// Copying from local source to Iota destination (upload)

		srcEx, err := homedir.Expand(src)
		if err != nil {
			return fmt.Errorf("invalid path %q", src)
		}

		// Check that the file exists
		info, err := os.Stat(srcEx)
		if err != nil {
			return fmt.Errorf("invalid path %q", src)
		}
		if info.IsDir() {
			// TODO: check that the --recursive flag is set
			return fmt.Errorf("%s is a directory. Please use the --recursive flag", src)
		}

		fmt.Printf("Uploading %s to %s\n", src, dst)

		f, err := os.Open(src)
		if err != nil {
			return fmt.Errorf("unable to open %s: %v", src, err)
		}
		err = client.UploadWithContext(c.Context, f, dst, iotafs.CompressZstd)
		if err != nil {
			return err
		}

	} else {
		// Local source & destination -- error
		return fmt.Errorf("at least one of <src> or <dst> must be an iota:// location")
	}

	return nil
}

func ls(client *iotafs.Client, c *cli.Context) error {
	args := c.Args()
	if args.Len() != 1 {
		return fmt.Errorf("only 1 argument expected")
	}

	pattern := args.Get(0)

	res, err := client.ListFiles(pattern)
	if err != nil {
		return err
	}

	for _, row := range res {
		ts := row.CreatedAt.Local().Format(time.RFC3339)
		fmt.Printf("%s  %s  %s\n", ts, humanBytes(row.Size), row.Name)
	}

	return nil
}

func humanBytes(size uint64) string {
	if size < 1024 {
		return fmt.Sprintf("%5.1d B  ", size)
	}
	s := float64(size) / 1024
	if s < 1024 {
		return fmt.Sprintf("%5.1f KiB", s)
	}
	s /= 1024
	if s < 1024 {
		return fmt.Sprintf("%5.1f MiB", s)
	}
	s /= 1024
	return fmt.Sprintf("%5.1f GiB", s)
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
			Usage:     "Copy a file to / from Iota",
			UsageText: "iota cp <src> <dst>",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "compression",
					Value: "zstd",
				},
			},
			Action: makeAction(cp),
		},
		{
			Name:      "ls",
			Usage:     "List files",
			UsageText: "iota ls <pattern>",
			Action:    makeAction(ls),
		},
	}

	app.Run(os.Args)
}
