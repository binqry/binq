package schema

import (
	"encoding/json"
	"fmt"

	"github.com/progrhyme/binq/internal/erron"
)

type IndiceItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Index struct {
	*indexProps
}

type indexProps struct {
	Items []IndiceItem `json:"items"`
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
