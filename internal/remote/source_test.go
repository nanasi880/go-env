package remote_test

import (
	"os"
	"testing"

	"github.com/nanasi880/go-env/internal/remote"
	"golang.org/x/net/html"
)

func TestVersion_Less(t *testing.T) {
	v1 := remote.Version{
		Major: 1,
		Minor: 5,
		Patch: 0,
		RC:    0,
		Beta:  0,
	}
	v2 := remote.Version{
		Major: 1,
		Minor: 3,
		Patch: 0,
		RC:    1,
		Beta:  0,
	}

	less := v1.Less(v2)
	t.Log(less)
}

func TestCrawlNode(t *testing.T) {
	f, err := os.Open("testdata/index.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	root, err := html.Parse(f)
	if err != nil {
		t.Fatal(err)
	}

	sources, err := remote.CrawlNode(root)
	_ = sources
}
