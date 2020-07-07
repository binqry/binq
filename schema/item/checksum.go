package item

import (
	"crypto/sha256"
	"hash"
	"hash/crc32"
	"strings"

	"github.com/progrhyme/go-lv"
)

type ChecksumType int

const (
	ChecksumTypeSHA256 ChecksumType = iota + 1
	ChecksumTypeCRC
	ChecksumTypeUnknown ChecksumType = -1
)

type ItemChecksum struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256,omitempty"`
	// CRC-32 IEEE Std.
	CRC string `json:"crc,omitempty"`
}

func NewItemChecksums(arg string) (sums []ItemChecksum) {
	if arg == "" {
		return nil
	}

	for _, entry := range strings.Split(arg, ",") {
		params := strings.Split(entry, ":")
		switch len(params) {
		case 2:
			sums = append(sums, ItemChecksum{File: params[0], SHA256: params[1]})
		case 3:
			switch params[2] {
			case "sha256", "SHA256", "SHA-256":
				sums = append(sums, ItemChecksum{File: params[0], SHA256: params[1]})
			case "crc", "CRC":
				sums = append(sums, ItemChecksum{File: params[0], CRC: params[1]})
			default:
				lv.Warnf("Unsupported algorithm: %s. Param: %s", params[2], entry)
			}

		default:
			lv.Warnf("Wrong argement for replacement: %s", entry)
		}
	}
	return sums
}

func (sum *ItemChecksum) GetSumAndHasher() (s string, h hash.Hash, t ChecksumType) {
	if sum.SHA256 != "" {
		return sum.SHA256, sha256.New(), ChecksumTypeSHA256
	} else if sum.CRC != "" {
		return sum.CRC, crc32.NewIEEE(), ChecksumTypeCRC
	}
	return "", nil, ChecksumTypeUnknown
}

func (sum *ItemChecksum) SetSum(val string, t ChecksumType) {
	switch t {
	case ChecksumTypeSHA256:
		sum.SHA256 = val
	case ChecksumTypeCRC:
		sum.SHA256 = val
	default:
		// Unexpected
		lv.Fatalf("Unsupported type for checksum: %d", t)
	}
}
