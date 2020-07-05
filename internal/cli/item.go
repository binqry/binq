package cli

import (
	"io/ioutil"

	"github.com/progrhyme/binq/internal/erron"
	"github.com/progrhyme/binq/schema/item"
)

func readAndDecodeItemJSONFile(file string) (raw []byte, obj *item.Item, err error) {
	raw, _err := ioutil.ReadFile(file)
	if _err != nil {
		err = erron.Errorwf(_err, "Can't read item file: %s", file)
		return raw, obj, err
	}
	obj, _err = item.DecodeItemJSON(raw)
	if _err != nil {
		err = erron.Errorwf(_err, "Failed to decode Item JSON: %s", file)
		return raw, obj, err
	}
	return raw, obj, nil
}
