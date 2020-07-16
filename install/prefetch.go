package install

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/binqry/binq/schema/item"
	"github.com/hashicorp/go-version"
)

// prefetch query metadata for item info to fetch
func (r *Runner) prefetch() (err error) {
	if strings.HasPrefix(r.Source, "http") {
		r.sourceURL = r.Source
		return nil
	}
	if r.ServerURL == nil {
		return fmt.Errorf("No server is configured. Can't deal with source: %s", r.Source)
	}

	name, tgtVer := parseSourceString(r.Source)
	tgt, _err := r.getClient().GetItemInfo(name)
	if _err != nil {
		return _err
	}

	var rev *item.ItemRevision
	if tgtVer == "" {
		rev = tgt.GetLatest()
	} else {
		rev = tgt.GetRevision(tgtVer)
	}
	if rev == nil {
		return fmt.Errorf("Version not found: %s", r.Source)
	}
	if ok, err := r.checkItemVersion(rev); !ok {
		return err
	}

	srcURL, err := rev.GetURL(item.FormatParam{OS: r.os, Arch: r.arch})
	if err != nil {
		return err
	}
	if srcURL == "" {
		return fmt.Errorf("Can't get source URL from JSON")
	}

	r.sourceURL = srcURL
	r.sourceItem = rev

	return nil
}

func (r *Runner) checkItemVersion(rev *item.ItemRevision) (ok bool, err error) {
	if r.NewerThan == "" {
		return true, nil
	}

	sbj, _err := version.NewVersion(rev.Version)
	if _err != nil {
		r.Logger.Warnf("Can't parse item's version %s as semantic. Continue installation...", rev.Version)
		return true, nil
	}
	threshold, _err := version.NewVersion(r.NewerThan)
	if _err != nil {
		r.Logger.Warnf("Can't parse given version %s as semantic. Continue installation...", r.NewerThan)
		return true, nil
	}
	if sbj.LessThanOrEqual(threshold) {
		r.Logger.Debugf("Item version %s <= %s. Stop installation.", rev.Version, r.NewerThan)
		return false, ErrVersionNotNewerThanThreshold
	}

	return true, nil
}

func parseSourceString(src string) (name, version string) {
	re := regexp.MustCompile(`^([\w\-\./]+)@([\w\-\.]+)$`)
	if re.MatchString(src) {
		matched := re.FindStringSubmatch(src)
		return matched[1], matched[2]
	}

	return src, ""
}
