#!/usr/bin/env bash
#
# Generate Release Note from CHANGELOG.md
# This script extracts content of specified VERSION in CHANGELOG.md using awk.

set -euo pipefail

VERSION=${VERSION:-}

if [[ -z "$VERSION" ]]; then
  VERSION="$(go run ./cmd/binq/*.go version | awk '{print $2}')"
fi

awk -v version="$VERSION" '
  BEGIN { st = 0 }
  {
    if (st == 0 && match($0, "^## " version)) {
      st = 1
      print
    } else if (st == 1 && /^##/) {
      exit
    } else if (st == 1) {
      print
    }
  }
' CHANGELOG.md
