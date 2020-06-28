# binq

Download items via HTTP and extract them when compressed.  
Mainly focuses on executable programs distributed on the internet.

# Install

Choose one of below methods:

- Download from GitHub releases
- go get (go command is required)

Description for each method follows.

## Download from GitHub Releases

Download latest binary from [GitHub Releases](https://github.com/progrhyme/binq/releases)
and put it under one directory in `$PATH` entries.

Typical commands to achieve this are following:

```sh
bin=/usr/local/bin  # Change to your favorite path
version=0.1.0       # Make sure this is the latest
os=darwin           # or "linux" is supported
tmpfile=$(mktemp)
curl -Lo $tmpfile "https://github.com/progrhyme/binq/releases/download/v${version}/binq_${version}_${os}_amd64.zip"
unzip -d $bin $tmpfile
rm $tmpfile
```

## go get

Just run this:

```sh
go get github.com/progrhyme/binq
```

# CLI Usage

Syntax:

```sh
# Download & Extract
binq SOURCE \
  [-d|--dir OUTPUT_DIR] [-f|--file OUTFILE] \
  [-s|--server SERVER] \
  [-v|--verbose] [--debug]

# Show help
binq -h|--help

# Show version
binq -V|--version
```

Examples:

```sh
# With full URL
binq https://github.com/peco/peco/releases/download/v0.5.7/peco_darwin_amd64.zip \
  -d path/to/bin
binq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 \
  -d path/to/bin -f jq

# With index server
binq -s https://progrhy.me/binq-index github.com/peco/peco -d path/to/bin
export BINQ_SERVER=https://progrhy.me/binq-index
binq jq -d path/to/bin -f jq
```

# Binq Index Server

`binq` refers to an index server to fetch meta data of an item when its identifier is specified
instead of full URL.  
We call it **Binq Index Server**.

It contains the database of downloadable items with their URLs for `binq`.  
When `binq` send a request to the server, it responds a JSON data which contains information about
the item.

A live example of index server is https://progrhy.me/binq-index/ .  
This is just a static site of GitHub Pages, whose source is https://github.com/progrhyme/binq-index/tree/gh-pages .

# License

The MIT License.

Copyright (c) 2020 IKEDA Kiyoshi.
