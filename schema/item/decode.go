package item

import (
	"encoding/json"

	"github.com/progrhyme/binq/internal/erron"
)

func DecodeItemJSON(b []byte) (item *Item, err error) {
	var i itemProps
	if _err := json.Unmarshal(b, &i); _err != nil {
		return item, erron.Errorwf(_err, "Failed to unmarshal JSON: %s", b)
	}
	item = &Item{itemProps: &i}
	return item, err
}
