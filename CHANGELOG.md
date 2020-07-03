## 0.4.0 (2020-07-03)

Feature: ([#4](https://github.com/progrhyme/binq/pull/4))

- (CLI) Add `new` subcommand to generate Item JSON
- (CLI) Add `revise` subcommand to add/edit/delete a Version in Item JSON
- (schema/item) Add functions to create/update/output Item data structure
- (schema/item) Support CRC checksum

Change: ([#4](https://github.com/progrhyme/binq/pull/4))

- (schema) Move Item-related functionality to "schema/item" subpackage
- (CLI) Add `version` subcommand and obsolete `-V|--version` option

## 0.3.1 (2020-07-01)

Feature: ([#3](https://github.com/progrhyme/binq/pull/3))

- (schema/item) Support `{{.BinExt}}` in "url-format" which is replaced with `.exe` on Windows, blank on others
- (schema/item) Support `{{.Ext}}` in "url-format" to customize file extension. Replacement for it is defined by "extension" hash in JSON

Change: ([#3](https://github.com/progrhyme/binq/pull/3))

- (client) Unexport type `CLI`

## 0.3.0 (2020-06-30)

Feature: ([#2](https://github.com/progrhyme/binq/pull/2))

- (index,query) Fallback to search INDEX to find ITEM when it first fails to fetch it from server
on the path specified as an argument

Bug Fix: ([#2](https://github.com/progrhyme/binq/pull/2))

- (CLI) Set mode = ModeDLOnly when both `--no-extract` & `--no-exec` options specified

Modify: ([#2](https://github.com/progrhyme/binq/pull/2))

- (client) Make client.RunOption visible so that Run is callable from outside

## 0.2.0 (2020-06-30)

Feature: ([#1](https://github.com/progrhyme/binq/pull/1))

- (CLI) Add `--no-extract|-z` option not to extract compressed archive
- (CLI) Add `--no-exec|-X` option not to take care of executable files

## 0.1.0 (2020-06-29)

Initial release.
