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
  BEGIN { current = 0 }
  {
    if (current == 0 && match($0, "^## " version)) {
      current = 1
      print $0
    } else if (current == 1 && /^##/) {
      current = -1
    } else if (current == 1) {
      print $0
    }
  }
' CHANGELOG.md
