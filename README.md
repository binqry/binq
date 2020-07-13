[![release](https://badgen.net/github/release/binqry/binq)](https://github.com/binqry/binq/releases)
[![go-test](https://github.com/binqry/binq/workflows/go-test/badge.svg)](https://github.com/binqry/binq/actions?query=workflow%3Ago-test)

# binq

Download items via HTTP and extract them when they are compressed in forms of zip, tar.gz etc.  
**binq** mainly focuses on executable programs distributed on the internet.

Typical use case is fetching GitHub release assets.  
But **binq** is not limited to it.

# Install

Choose one of below methods:

- [Homebrew](https://brew.sh/) or [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux) (using Tap)
- Download from GitHub releases
- go get (go command is required)

Description for each method follows.

## Homebrew (Linuxbrew)

```sh
brew tap progrhyme/taps
brew install binq
```

## Download from GitHub Releases

Download latest binary from [GitHub Releases](https://github.com/binqry/binq/releases)
and put it under one directory in `$PATH` entries.

Typical commands to achieve this are following:

```sh
bin=/usr/local/bin  # Change to your favorite path
version=0.6.3       # Make sure this is the latest
os=darwin           # or "linux" is supported
tmpfile=$(mktemp)
curl -Lo $tmpfile "https://github.com/binqry/binq/releases/download/v${version}/binq_${version}_${os}_amd64.zip"
unzip -d $bin $tmpfile
rm $tmpfile
```

## go get

Just run this:

```sh
go get github.com/binqry/binq/cmd/binq
```

# CLI Usage

Syntax:

```sh
# Main Command. Download & Extract target binary/archive
binq [install] [-t|--target] SOURCE[@VERSION] \
  [-d|--dir OUTPUT_DIR] [-f|--file OUTFILE] \
  [-s|--server SERVER] \
  [-z|--no-extract] [-X|--no-exec] \
  [GENERAL_OPTIONS]

# Other Commands
binq new         # Create Item JSON for Index Server
binq revise      # Add/Edit/Delete a version in Item JSON
binq verify      # Verify checksum of a version in item JSON
binq register    # Register item JSON into Local Index Dataset
binq modify      # Modify item properties on Local Index
binq deregister  # Deregister item from Local Index Dataset
binq version     # Show binq version

# Show help
binq [COMMAND] -h|--help
```

General options for all commands:

```
-h|--help               # Show help
-L, --log-level string  # Log level (debug,info,notice,warn,error)
```

## binq install

Examples:

```sh
# With full URL
binq https://github.com/peco/peco/releases/download/v0.5.7/peco_darwin_amd64.zip \
  -d path/to/bin
binq https://github.com/stedolan/jq/releases/download/jq-1.6/jq-linux64 \
  -d path/to/bin -f jq

# With Index Server
binq -s https://binqry.github.io/index/ mdbook -d path/to/bin
export BINQ_SERVER=https://binqry.github.io/index/
export BINQ_BIN_DIR=path/to/bin
binq jq@1.6
```

[Index Server](#binq-index-server) serves meta data of downloadable items by **binq**.  
See following section for more details.

Command specific options for `binq install`:

```
-d, --directory string   # Output Directory
-f, --file string        # Output File name
-X, --no-exec            # Don't care for executable files
-z, --no-extract         # Don't extract archive
-s, --server string      # Index Server URL
-t, --target string      # Target Item (Name or URL)
```

## Manipulate Item JSON

**binq** has some commands to create/edit **Item JSON** for [Binq Index Server](#binq-index-server).  
Each Item JSON represents a manifest to download & install item by `binq` command.  
It includes followings:

- **Download URL format** to determine the URL for specific version, OS or architecture
- **Versions** available for download
- **Checksums** to verify downloaded items

Commands Syntax:

```sh
# Generate Item JSON
binq new URL_FORMAT [-v|--version VERSION] [-f|--file OUTPUT_FILE] \
  [-r|--replace REPLACEMENTS] [-e|--ext EXTENSIONS] [-R|--rename RENAME_FILES] [GENERAL_OPTIONS]

# Add or Update Version in Item JSON
binq revise ITEM_JSON_FILE [-v|--version] VERSION \
  [-s|--sum CHECKSUMS] [-u|--url URL_FORMAT] [-r|--replace REPLACEMENTS] [-e|--ext EXTENSIONS] \
  [-R|--rename RENAME_FILES] [--latest] [--no-latest] [-y|--yes] [GENERAL_OPTIONS]

# Download a Version in Item JSON and Verify its checksum
binq verify path/to/item.json [-v|--version VERSION] [-o|--os OS] [-a|--arch ARCH] \
  [-y|--yes] [--keep] [GENERAL_OPTIONS]

# Delete Version in Item JSON
binq revise ITEM_JSON_FILE VERSION --delete [-y|--yes] [GENERAL_OPTIONS]
```

## Manipulate Local Index Dataset

You can interact with Index Dataset using `binq` command, but currently CLI only supports Local
Dataset in filesystem.

Commands Syntax:

```sh
# Register or Update Item content on Local Index Dataset
binq register pato/to/root[/index.json] path/to/item.json \
  [-n|--name NAME] [-p|--path PATH] [-y|--yes] [GENERAL_OPTIONS]

# Modify Item properties in Local Index Dataset
binq modify pato/to/root[/index.json] NAME \
  [-n|--name NEW_NAME] [-p|--path PATH] [-y|--yes] [GENERAL_OPTIONS]

# Deregister Item from Local Index Dataset
binq deregister pato/to/root[/index.json] NAME [-y|--yes] [GENERAL_OPTIONS]
```

# Binq Index Server

**binq** can refer to an index server to fetch meta data of an item when its identifier is specified
instead of full URL.  
We call it **Binq Index Server**.

It contains the dataset of items with their downloadable URLs.  
When **binq** send a request to the server, it responds a JSON data which contains information of
the item.

A live example of index server is https://binqry.github.io/index/ .  
This is just a static site of GitHub Pages, whose source is https://github.com/binqry/index/tree/gh-pages .

# License

The MIT License.

Copyright (c) 2020 IKEDA Kiyoshi.
