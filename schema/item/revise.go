package item

import (
	"github.com/hashicorp/go-version"
	"github.com/progrhyme/binq/internal/logs"
)

type ReviseMode int

const (
	ReviseModeNatural ReviseMode = iota + 1
	ReviseModeLatest
	ReviseModeOld
)

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
		logs.Fatalf("Fatal Error! Invalid mode: %v", mode)
	}
}

func (i *Item) addNotLatestRevision(rev *ItemRevision, mode ReviseMode) {
	newVer, err := version.NewVersion(rev.Version)
	if err != nil {
		logs.Noticef("Given version is not parsed as version")
		for idx, rv := range i.Versions {
			_, err := version.NewVersion(rv.Version)
			if err != nil {
				logs.Debugf("Fail to parse version: %s", rv.Version)
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
			logs.Noticef("Cannot parse version: %s", rv.Version)
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
