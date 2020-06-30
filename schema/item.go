package schema

import (
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	"github.com/progrhyme/binq/internal/erron"
)

// Item wraps itemProps which corresponds to JSON structure of item data
type Item struct {
	*itemProps
}

type ItemRevision struct {
	Version      string            `json:"version"`
	Checksums    []itemChecksums   `json:"checksums,omitempty"`
	URLFormat    string            `json:"url-format,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Extension    map[string]string `json:"extension,omitempty"`
}

type ItemURLParam struct {
	Version string
	OS      string
	Arch    string
}

type itemChecksums struct {
	File   string `json:"file"`
	Sha256 string `json:"sha256"`
}

type itemProps struct {
	Meta struct {
		URLFormat    string            `json:"url-format,omitempty"`
		Replacements map[string]string `json:"replacements,omitempty"`
		Extension    map[string]string `json:"extension,omitempty"`
	} `json:"meta,omitempty"`
	Latest struct {
		Version string `json:"version"`
	} `json:"latest,omitempty"`
	Versions []ItemRevision `json:"versions,omitempty"`
}

func DecodeItemJSON(b []byte) (item *Item, err error) {
	var i itemProps
	if _err := json.Unmarshal(b, &i); _err != nil {
		return item, erron.Errorwf(_err, "Failed to unmarshal JSON: %s", b)
	}
	item = &Item{itemProps: &i}
	return item, err
}

func (i *Item) String() string {
	return fmt.Sprintf("%+v", *i.itemProps)
}

func (i *Item) GetLatestURL(param ItemURLParam) (url string, err error) {
	rev := i.GetLatest()
	if rev == nil {
		return "", nil
	}
	return rev.GetURL(param)
}

func (i *Item) GetLatest() (rev *ItemRevision) {
	latest := i.Latest
	if latest.Version == "" {
		return nil
	}

	found := i.GetRevision(latest.Version)
	if found != nil {
		return found
	}

	return &ItemRevision{
		Version:      latest.Version,
		URLFormat:    i.Meta.URLFormat,
		Replacements: i.Meta.Replacements,
		Extension:    i.Meta.Extension,
	}
}

func (i *Item) GetRevision(version string) (rev *ItemRevision) {
	tmp := &ItemRevision{
		Version:      version,
		URLFormat:    i.Meta.URLFormat,
		Replacements: i.Meta.Replacements,
		Extension:    i.Meta.Extension,
	}

	found := false
	for _, ver := range i.Versions {
		if ver.Version == version {
			found = true
			tmp.Checksums = ver.Checksums
			if ver.URLFormat != "" {
				tmp.URLFormat = ver.URLFormat
			}
			if ver.Replacements != nil {
				tmp.Replacements = ver.Replacements
			}
			if ver.Extension != nil {
				tmp.Extension = ver.Extension
			}
			break
		}
	}
	if found {
		return tmp
	}

	return nil
}

func (rev *ItemRevision) GetURL(param ItemURLParam) (url string, err error) {
	// Convert param into map to apply replacements
	hash := make(map[string]string)
	hash["Version"] = rev.Version
	hash["OS"] = param.OS
	hash["Arch"] = param.Arch
	if param.OS == "windows" {
		hash["BinExt"] = ".exe"
	}
	if rev.Extension != nil {
		if ext, ok := rev.Extension[param.OS]; ok {
			hash["Ext"] = ext
		} else {
			hash["Ext"] = rev.Extension["default"]
		}
	}

	replaced := make(map[string]string)
	for key, val := range hash {
		for orig, rep := range rev.Replacements {
			if val == orig {
				replaced[key] = rep
				break
			}
		}
		if replaced[key] == "" {
			replaced[key] = val
		}
	}

	var b strings.Builder
	t := template.Must(template.New("url").Parse(rev.URLFormat))

	if _err := t.Execute(&b, replaced); _err != nil {
		err = erron.Errorwf(
			_err, "Failed to exec template. Format: %s, Params: %v", rev.URLFormat, replaced)
		return "", err
	}

	return b.String(), nil
}
