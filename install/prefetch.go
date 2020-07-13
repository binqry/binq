package install

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/binqry/binq/schema/item"
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

	name, version := parseSourceString(r.Source)
	tgt, _err := r.getClient().GetItemInfo(name)
	if _err != nil {
		return _err
	}

	var rev *item.ItemRevision
	if version == "" {
		rev = tgt.GetLatest()
	} else {
		rev = tgt.GetRevision(version)
	}
	if rev == nil {
		return fmt.Errorf("Version not found: %s", r.Source)
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

func parseSourceString(src string) (name, version string) {
	re := regexp.MustCompile(`^([\w\-\./]+)@([\w\-\.]+)$`)
	if re.MatchString(src) {
		matched := re.FindStringSubmatch(src)
		return matched[1], matched[2]
	}

	return src, ""
}
