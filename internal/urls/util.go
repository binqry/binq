package urls

import (
	"net/url"
	"path"

	"github.com/binqry/binq/internal/erron"
)

func AddPath(obj *url.URL, pth string) {
	obj.Path = path.Join(obj.Path, pth)
}

func Join(addr, pth string) (joined string, err error) {
	obj, _err := url.Parse(addr)
	if _err != nil {
		return "", erron.Errorwf(_err, "URL parse failed.")
	}
	AddPath(obj, pth)
	return obj.String(), nil
}
