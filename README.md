[![release](https://badgen.net/github/release/binqry/binq)](https://github.com/binqry/binq/releases)
[![go-test](https://github.com/binqry/binq/workflows/go-test/badge.svg)](https://github.com/binqry/binq/actions?query=workflow%3Ago-test)

# binq

**binq** is a light-weight software installer written in Golang.  
It downloads stuff via HTTP and extracts them when they are compressed in forms of zip, tar.gz etc.  
This tool mainly focuses on executable programs distributed on the internet; and makes it easier
when they are not provided by any package manager.

Typical use case is fetching GitHub release assets.  
But **binq** is not limited to it.

# Documentation

https://binqry.github.io/

# System Requirements

Pre-built binaries are available for Windows, macOS and Linux with x86-64 CPU architecture.

**binq** is logically supposed to work on any machine for which Go can compile codes.

# Install

Choose one of below methods:

- [Homebrew](https://brew.sh/) or [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux) (using Tap)
- Download from GitHub releases
- go get (go command is required)

Description for each method follows.

## Homebrew (Linuxbrew)

```sh
brew tap progrhyme/tap
brew install binq
```

## Download from GitHub Releases

Download latest binary from [GitHub Releases](https://github.com/binqry/binq/releases)
and put it under one directory in `$PATH` entries.

The following script detects your OS and downloads the latest binq binary into your current directory:

```sh
curl -s "https://raw.githubusercontent.com/binqry/binq/master/get-binq.sh" | bash
```

## go get

Just run this:

```sh
go get github.com/binqry/binq/cmd/binq
```

# Usage at a Glance

Examples:

```sh
binq mdbook
binq jq@1.6 -d /usr/local/bin
```

These commands will install `mdbook` or `jq` binary.  
If `-d|--dir` option is not provided, it will be installed at current directory.

Other ways to install stuff:

```sh
binq https://github.com/peco/peco/releases/download/v0.5.7/peco_darwin_amd64.zip \
  -d path/to/bin

export BINQ_BIN_DIR=path/to/bin
binq kustomize
```

Other commands:

```sh
binq index         # List Items on Index Server
binq self-upgrade  # Upgrade binq binary itself
binq new           # Create Item Manifest
binq revise        # Add/Edit/Delete a version in Item Manifest
binq verify        # Verify checksum of a version in Item Manifest
binq register      # Register or Update Item Manifest onto Local Index Dataset
binq modify        # Modify Item properties on Local Index Dataset
binq deregister    # Deregister Item from Local Index
binq version       # Show binq version

# Show help
binq [COMMAND] -h|--help
```

See the [documentation](https://binqry.github.io/) for details and more information.

# License

The MIT License.

Copyright (c) 2020 IKEDA Kiyoshi.
