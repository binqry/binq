## 0.6.3 (2020-07-12)

Features: ([#9](https://github.com/progrhyme/binq/pull/9))

- (schema,install) Add ".meta.rename-files" property to Item JSON to specify default file-renaming config
- (CLI/new,revise) Add option `-R|--rename RENAME_FILES` to add "rename-files" JSON property

Minor Change: ([#9](https://github.com/progrhyme/binq/pull/9))

- (schema/item) Rename `item.ItemURLParam` to `FormatParam`

Other: ([#9](https://github.com/progrhyme/binq/pull/9))

- (test)(CLI/new,revise) Add some tests

## 0.6.2 (2020-07-12)

Enhance: ([#8](https://github.com/progrhyme/binq/pull/8))

- (checksum) Support MD5 checksum defined in RFC 1321

Bug Fix: ([#8](https://github.com/progrhyme/binq/pull/8))

- (checksum) `func (*item.ItemChecksum) SetSum` wrongly sets SHA256 value when CRC is expected

## 0.6.1 (2020-07-11)

Feature: ([#7](https://github.com/progrhyme/binq/pull/7))

- (CLI/install) Enable to specify download version by `@VERSION` suffix to source target argument
- (CLI/install) Enable to specify default install directory by `$BINQ_BIN_DIR` environment variable

Bug Fix: ([#7](https://github.com/progrhyme/binq/pull/7))

- (CLI/new) Can't create a file with `-f|--file` option

Other: ([#7](https://github.com/progrhyme/binq/pull/7))

- (internal) Use github.com/progrhyme/go-lv package as logger instead of `internal/logs`

## 0.6.0 (2020-07-06)

Feature: ([#6](https://github.com/progrhyme/binq/pull/6))

- (CLI) Add `verify` subcommand to verify checksum of a Version in Item JSON

Change: ([#6](https://github.com/progrhyme/binq/pull/6))

- (CLI) Set logging level by `-L|--log-level LEVEL` option. Obsolete `--debug` & `-v|--verbose` options
- (client) Export func: NewHttpClient & NewHttpGetRequest

Minor Fix: ([#6](https://github.com/progrhyme/binq/pull/6))

- (client) Fix potential bug: wrong condition to clear tempdir on downloading item

Other: ([#6](https://github.com/progrhyme/binq/pull/6))

- (testing,CI) Add some CLI tests & CI task for testing

## 0.5.0 (2020-07-04)

Feature: ([#5](https://github.com/progrhyme/binq/pull/5))

- (CLI) Add `register` subcommand to add Item into Index
- (CLI) Add `deregister` subcommand to remove Item from Index
- (CLI) Add `modify` subcommand to edit properties of Item in Index
- (install, client) Verify checksum of downloaded item when it is provided by Item JSON

Change: ([#5](https://github.com/progrhyme/binq/pull/5))

- (schema/item) Rename struct ItemChecksums -> ItemChecksum

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
