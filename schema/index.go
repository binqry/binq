package schema

import (
	"encoding/json"
	"fmt"

	"github.com/progrhyme/binq/internal/erron"
)

type Index struct {
	*indexProps
}

type indexProps struct {
	Items []IndiceItem `json:"items"`
}

func NewIndex() (idx *Index) {
	return &Index{&indexProps{Items: []IndiceItem{}}}
}

func DecodeIndexJSON(b []byte) (index *Index, err error) {
	var ip indexProps
	if _err := json.Unmarshal(b, &ip); _err != nil {
		return index, erron.Errorwf(_err, "Failed to unmarshal JSON: %s", b)
	}
	index = &Index{indexProps: &ip}
	return index, err
}

func (idx *Index) String() string {
	return fmt.Sprintf("%+v", *idx.indexProps)
}

func (idx *Index) Print(pretty bool) (b []byte, err error) {
	var _err error
	if pretty {
		b, _err = json.MarshalIndent(idx, "", "  ")
	} else {
		b, _err = json.Marshal(idx)
	}
	if _err != nil {
		return b, erron.Errorwf(_err, "Failed to marshal JSON: %s", idx)
	}
	return b, nil
}

func (idx *Index) Find(name string) (indice *IndiceItem) {
	for _, i := range idx.Items {
		if i.Name == name {
			return &i
		}
	}
	return nil
}

func (idx *Index) FindPath(name string) (path string) {
	if item := idx.Find(name); item != nil {
		return item.Path
	}
	return ""
}

func (idx *Index) Add(indice *IndiceItem) {
	for i, entry := range idx.Items {
		if entry.Name > indice.Name {
			idx.Items = append(idx.Items[:i], append([]IndiceItem{*indice}, idx.Items[i:]...)...)
			return
		}
	}
	idx.Items = append(idx.Items, *indice)
}

func (idx *Index) Swap(name string, indice *IndiceItem) (success bool) {
	for i, entry := range idx.Items {
		if entry.Name == name {
			idx.Items = append(idx.Items[:i], append([]IndiceItem{*indice}, idx.Items[i+1:]...)...)
			return true
		}
	}
	return false
}
