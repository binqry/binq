package schema

import (
	"fmt"
)

type IndiceItem struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func (i *IndiceItem) String() (s string) {
	return fmt.Sprintf(`{"name":"%s", "path":"%s"}`, i.Name, i.Path)
}
