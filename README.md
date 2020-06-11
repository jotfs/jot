# Jot

[![Docs](https://godoc.org/github.com/jotfs/jot?status.svg)](https://pkg.go.dev/github.com/jotfs/jot?tab=doc) [![Build Status](https://travis-ci.org/jotfs/jot.svg?branch=master)](https://travis-ci.org/jotfs/jot) [![Go Report Card](https://goreportcard.com/badge/github.com/jotfs/jot)](https://goreportcard.com/report/github.com/jotfs/jot)

The Go client library for [JotFS](https://github.com/jotfs/jotfs), and associated CLI `jot`.

## Library install

```
go get -u github.com/jotfs/jot
```

## CLI install

### Binaries

Download the latest binary from the [releases](https://github.com/jotfs/jot/releases) page and extract. Example:

```
gzip -dc ./jot_darwin_amd64_v0.0.2.gz > jot
chmod +x jot
./jot --help
```

### Source

`jot` may also be installed from source:
```
git clone https://github.com/jotfs/jot.git
cd jot
go install ./cmd/jot
jot --help
```

## CLI configuration

The URL of your JotFS server may be passed using the `--endpoint` option. Alternatively, a config file may used with location at `$HOME/.jot/config.toml` or by setting the `JOT_CONFIG_FILE` environment variable.

Example `config.toml`:
```
[[profile]]
name = "default"
endpoint = "http://localhost:6777"

[[profile]]
name = "prod"
endpoint = "https://example.com
```

By default, `jot` will look for a profile named `default`. This may be overridden by setting the `--profile` option.

## CLI reference

```
NAME:
   jot - A CLI tool for working with a JotFS server

USAGE:
   jot [global options] command [command options] [arguments...]

DESCRIPTION:
   
   jot will look for its configuration file at $HOME/.jot/config.toml 
   by default. Alternatively, its path may be specified by setting the 
   JOT_CONFIG_FILE environment variable, or with the --config option.
   The server endpoint URL may be overridden with the --endpoint option.

COMMANDS:
   cp       copy files to / from JotFS
   ls       list files
   rm       remove files
   admin    server administration commands
   version  output version info
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value     path to config file
   --profile value    config profile to use (default: "default")
   --endpoint value   set the server endpoint
   --tls_skip_verify  skip TLS certificate verification (default: false)
   --help, -h         show help (default: false)
```


## License

Jot is licensed under the Apache 2.0 License. See [LICENSE](./LICENSE) for details.


