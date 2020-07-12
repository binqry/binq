package item

type ReviseMode int

const (
	ReviseModeNatural ReviseMode = iota + 1
	ReviseModeLatest
	ReviseModeOld
)

type FormatParam struct {
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
	RenameFiles  map[string]string `json:"rename-files,omitempty"`
}

type itemLatestRevision struct {
	Version string `json:"version"`
}
