package item

import (
	"encoding/json"

	"github.com/progrhyme/binq/internal/erron"
)

func GenerateItemJSON(rev *ItemRevision, pretty bool) (b []byte, err error) {
	var _err error
	prop := itemProps{
		Meta: itemMeta{
			URLFormat:    rev.URLFormat,
			Replacements: rev.Replacements,
			Extension:    rev.Extension,
		},
		Latest: itemLatestRevision{Version: rev.Version},
		Versions: []ItemRevision{
			{Version: rev.Version},
		},
	}
	if pretty {
		b, _err = json.MarshalIndent(prop, "", "  ")
	} else {
		b, _err = json.Marshal(prop)
	}
	if _err != nil {
		return b, erron.Errorwf(_err, "Failed to marshal JSON: %+v", rev)
	}
	return b, nil
}

func (i *Item) Print(pretty bool) (b []byte, err error) {
	var _err error
	if pretty {
		b, _err = json.MarshalIndent(i, "", "  ")
	} else {
		b, _err = json.Marshal(i)
	}
	if _err != nil {
		return b, erron.Errorwf(_err, "Failed to marshal JSON: %s", i)
	}
	return b, nil
}
