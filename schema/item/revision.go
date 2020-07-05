package item

import (
	"html/template"
	"strings"

	"github.com/progrhyme/binq/internal/erron"
)

type ItemRevision struct {
	Version      string            `json:"version"`
	Checksums    []ItemChecksum    `json:"checksums,omitempty"`
	URLFormat    string            `json:"url-format,omitempty"`
	Replacements map[string]string `json:"replacements,omitempty"`
	Extension    map[string]string `json:"extension,omitempty"`
}

func (rev *ItemRevision) GetChecksum(file string) (sum *ItemChecksum) {
	for _, cs := range rev.Checksums {
		if cs.File == file {
			return &cs
		}
	}
	return nil
}

func (rev *ItemRevision) AddOrSwapChecksum(sum *ItemChecksum) {
	for i, cs := range rev.Checksums {
		if cs.File == sum.File {
			rev.Checksums[i] = *sum
			return
		}
	}
	rev.Checksums = append(rev.Checksums, *sum)
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
