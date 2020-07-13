package urls

import (
	"net/url"
	"testing"
)

func TestAddPath(t *testing.T) {
	addr := "https://example.com/"
	u, err := url.Parse(addr)
	if err != nil {
		t.Errorf("URL parse failed. %v", err)
	}
	AddPath(u, "foo")
	want := "https://example.com/foo"
	if u.String() != want {
		t.Errorf("URL does not match. Want: %s, Got: %s", want, u.String())
	}
}

func TestJoin(t *testing.T) {
	addr := "https://example.com/"
	joined, err := Join(addr, "foo")
	if err != nil {
		t.Errorf("URL join failed. %v", err)
	}
	want := "https://example.com/foo"
	if joined != want {
		t.Errorf("URL does not match. Want: %s, Got: %s", want, joined)
	}

}
