package schema

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/binqry/binq/internal/erron"
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

func (idx *Index) ToJSON(pretty bool) (b []byte, err error) {
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

func (idx *Index) ToText() (text string) {
	format := "%-16s    %-48s"
	a := make([]string, len(idx.Items)+1)
	a = []string{fmt.Sprintf(format, "Name", "Path")}
	a = append(a, fmt.Sprint(strings.Repeat("=", 68)))
	for _, i := range idx.Items {
		a = append(a, fmt.Sprintf(format, i.Name, i.Path))
	}
	return strings.Join(a, "\n") + "\n"
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

// Add adds indice into idx. If an IndiceItem in idx.Items has the same name to indice, it is
// returned as conflictEntry
func (idx *Index) Add(indice *IndiceItem) (conflictEntry *IndiceItem) {
	j := -1
	for i, entry := range idx.Items {
		if entry.Name == indice.Name {
			return &entry
		}
		if j == -1 && entry.Name > indice.Name {
			j = i
		}
	}
	if j >= 0 {
		idx.Items = append(idx.Items[:j], append([]IndiceItem{*indice}, idx.Items[j:]...)...)
	} else {
		idx.Items = append(idx.Items, *indice)
	}
	return nil
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

func (idx *Index) Remove(name string) (success bool) {
	for i, entry := range idx.Items {
		if entry.Name == name {
			idx.Items = append(idx.Items[:i], idx.Items[i+1:]...)
			return true
		}
	}
	return false
}
