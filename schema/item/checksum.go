package item

import (
	"strings"

	"github.com/progrhyme/binq/internal/logs"
)

type ItemChecksums struct {
	File   string `json:"file"`
	Sha256 string `json:"sha256"`
}

func NewItemChecksums(arg string) (sums []ItemChecksums) {
	if arg == "" {
		return nil
	}

	for _, kv := range strings.Split(arg, ",") {
		var k, v string
		for idx, s := range strings.Split(kv, ":") {
			switch idx {
			case 0:
				k = s
			case 1:
				v = s
			default:
				logs.Warnf("Wrong argement for replacement: %s", kv)
				break
			}
		}
		sums = append(sums, ItemChecksums{File: k, Sha256: v})
	}
	return sums
}
