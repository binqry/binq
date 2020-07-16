#!/usr/bin/env bash
#
# get-binq.sh
#
# This script downloads and installs the latest binq binary to your current working directory;
# fails if the file "binq" already exists.

set -euo pipefail

if [[ -f binq ]]; then
  echo "\"binq\" already exists in this directory. Remove it first." >&2
  exit 1
fi

tmpfile=$(mktemp)
if [[ ! -r "$tmpfile" ]]; then
  echo "Can't make tmpfile. Quit." >&2
  exit 1
fi

trap "rm -f \"$tmpfile\"" EXIT

os=darwin
if [[ "$OSTYPE" =~ ^linux.*$ ]]; then
  os=linux
fi

echo "Download latest binq archive ..."
curl -s https://api.github.com/repos/binqry/binq/releases \
  | grep browser_download \
  | awk -v os=$os 'archive = "binq_.*" os "_amd64.zip"; $2 ~ archive { print $2 }' \
  | sort | tail -n 1 \
  | xargs curl -Lo $tmpfile

unzip -d . $tmpfile

./binq version

echo
echo "binq is installed in current directory."

exit 0
