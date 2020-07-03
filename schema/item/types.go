package item

import (
	"fmt"
)

// Item wraps itemProps which corresponds to JSON structure of item data
type Item struct {
	*itemProps
}

func (i *Item) String() string {
	return fmt.Sprintf("%+v", *i.itemProps)
}

type ItemRevision struct {
	Version      string            `json:"version"`
	Checksums    []ItemChecksums   `json:"checksums,omitempty"`
	URLFormat    string            `json:"url-format,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Extension    map[string]string `json:"extension,omitempty"`
}

type ItemURLParam struct {
	Version string
	OS      string
	Arch    string
}

// itemProps represents actual structure of Item JSON
type itemProps struct {
	Meta     itemMeta           `json:"meta,omitempty"`
	Latest   itemLatestRevision `json:"latest,omitempty"`
	Versions []ItemRevision     `json:"versions,omitempty"`
}

type itemMeta struct {
	URLFormat    string            `json:"url-format,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Extension    map[string]string `json:"extension,omitempty"`
}

type itemLatestRevision struct {
	Version string `json:"version"`
}
