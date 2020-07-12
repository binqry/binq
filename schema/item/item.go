package item

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/progrhyme/binq/internal/erron"
	"github.com/progrhyme/go-lv"
)

// Item wraps itemProps which corresponds to JSON structure of item data
type Item struct {
	*itemProps
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

func (i *Item) GetLatestURL(param FormatParam) (url string, err error) {
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
		RenameFiles:  i.Meta.RenameFiles,
	}
}

func (i *Item) GetRevision(version string) (rev *ItemRevision) {
	tmp := &ItemRevision{
		Version:      version,
		URLFormat:    i.Meta.URLFormat,
		Replacements: i.Meta.Replacements,
		Extension:    i.Meta.Extension,
		RenameFiles:  i.Meta.RenameFiles,
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
			if ver.RenameFiles != nil {
				tmp.Extension = ver.RenameFiles
			}
			break
		}
	}
	if found {
		return tmp
	}

	return nil
}

func (i *Item) AddOrUpdateRevision(rev *ItemRevision, mode ReviseMode) {
	var replaced bool
	for idx, rv := range i.Versions {
		if rv.Version == rev.Version {
			i.Versions[idx] = *rev
			replaced = true
			break
		}
	}

	switch mode {
	case ReviseModeLatest:
		i.Latest = itemLatestRevision{Version: rev.Version}
		if !replaced {
			i.Versions = append([]ItemRevision{*rev}, i.Versions...)
		}

	case ReviseModeOld, ReviseModeNatural:
		if replaced {
			break
		}
		i.addNotLatestRevision(rev, mode)

	default:
		lv.Fatalf("Fatal Error! Invalid mode: %v", mode)
	}
}

func (i *Item) UpdateRevisionChecksum(ver string, sum *ItemChecksum) (success bool) {
	for idx, rev := range i.Versions {
		if rev.Version == ver {
			rev.AddOrSwapChecksum(sum)
			i.Versions[idx] = rev
			return true
		}
	}
	return false
}

func (i *Item) addNotLatestRevision(rev *ItemRevision, mode ReviseMode) {
	newVer, err := version.NewVersion(rev.Version)
	if err != nil {
		lv.Noticef("Given version is not parsed as version")
		for idx, rv := range i.Versions {
			_, err := version.NewVersion(rv.Version)
			if err != nil {
				lv.Debugf("Fail to parse version: %s", rv.Version)
				// Insert before rv
				i.Versions = append(i.Versions[:idx], append([]ItemRevision{*rev}, i.Versions[idx:]...)...)
				if idx == 0 {
					i.Latest = itemLatestRevision{Version: rev.Version}
				}
				return
			}
		}
		// Push
		i.Versions = append(i.Versions, *rev)
		return
	}

	for idx := len(i.Versions) - 1; idx >= 0; idx-- {
		rv := i.Versions[idx]
		v, err := version.NewVersion(rv.Version)
		if err != nil {
			lv.Noticef("Cannot parse version: %s", rv.Version)
			continue
		}
		if v.GreaterThan(newVer) {
			// Insert After rv
			i.Versions = append(i.Versions[:idx+1], append([]ItemRevision{*rev}, i.Versions[idx+1:]...)...)
			return
		}
	}

	// Largest Version
	i.Versions = append([]ItemRevision{*rev}, i.Versions...)
	if mode == ReviseModeNatural {
		i.Latest = itemLatestRevision{Version: rev.Version}
	}
}

func (i *Item) DeleteRevision(version string) (deleted bool) {
	var latest bool
	var deletedIdx int

	if version == i.Latest.Version {
		latest = true
		deleted = true
	}
	for idx, rev := range i.Versions {
		if version == rev.Version {
			deleted = true
			deletedIdx = idx
			continue
		}
	}

	if !deleted {
		return false
	}

	if len(i.Versions) > 0 {
		i.Versions = append(i.Versions[:deletedIdx], i.Versions[deletedIdx+1:]...)
	}

	if !latest {
		return true
	}

	if len(i.Versions) > 0 {
		i.Latest = itemLatestRevision{Version: i.Versions[0].Version}
	} else {
		i.Latest = itemLatestRevision{}
	}

	return true
}
