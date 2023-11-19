package remote

import (
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/net/html"
)

const baseURL = "https://go.dev"

// Source はソースコードのメタデータです。
type Source struct {
	Version Version
	HRef    string
	Size    int64
	Hash    []byte
}

// GetHashString はハッシュ値を文字列表現で取得します。
func (source Source) GetHashString() string {
	return hex.EncodeToString(source.Hash)
}

// Download は指定されたディレクトリにソースコードをダウンロードします。
func (source Source) Download(dir string) (Archive, error) {
	resp, err := http.Get(baseURL + source.HRef)
	if err != nil {
		return Archive{}, errors.Wrapf(err, "http.Get returned error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return Archive{}, errors.Newf("Server returned a non 200 status code: %d", resp.StatusCode)
	}

	full := filepath.Join(dir, source.Version.String()+".src.tar.gz")

	// download
	err = func() error {
		f, err := os.OpenFile(full, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
		if err != nil {
			return errors.Wrap(err, "")
		}
		defer f.Close()

		_, err = io.Copy(f, resp.Body)
		if err != nil {
			return errors.Wrap(err, "")
		}

		return nil
	}()
	if err != nil {
		return Archive{}, err
	}

	return Archive{
		fullName: full,
		source:   source,
	}, nil
}

func Crawl() ([]Source, error) {
	resp, err := http.Get(baseURL + "/dl/")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.Newf("Server returned a non 200 status code: %d", resp.StatusCode)
	}

	root, err := html.Parse(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "html.Parse error")
	}

	return crawl(root)
}

func CrawlNode(root *html.Node) ([]Source, error) {
	return crawl(root)
}

func crawl(root *html.Node) ([]Source, error) {
	// find <body> tag
	body := findFirstNodeByTag(root, "body")
	if body == nil {
		return nil, errors.New("missing <body> tag")
	}

	// find <table class="downloadtable"> tag
	tables := findNodeByTagAndClass(body, "table", "downloadtable")

	// filter <tr> tag
	var tableElements []*html.Node
	for _, table := range tables {
		e := findNodeByTag(table, "tr")
		tableElements = append(tableElements, e...)
	}

	result := make([]Source, 0)
	for _, elem := range tableElements {
		td := findNodeByTag(elem, "td")
		if len(td) != 6 {
			continue
		}

		// only Source type
		if td[1].FirstChild == nil {
			continue
		}
		td1 := td[1].FirstChild
		if td1.Type != html.TextNode || td1.Data != "Source" {
			continue
		}

		if hasAttrValue(td[0], "class", "filename") == false {
			continue
		}
		a := findFirstNodeByTagAndClass(td[0], "a", "download")
		if a == nil || a.FirstChild == nil || a.FirstChild.Type != html.TextNode {
			continue
		}

		if td[4].FirstChild == nil || td[4].FirstChild.Type != html.TextNode {
			continue
		}

		tt := findFirstNodeByTag(td[5], "tt")
		if tt == nil || tt.FirstChild == nil || tt.FirstChild.Type != html.TextNode {
			continue
		}

		version, err := parseVersion(strings.TrimSuffix(a.FirstChild.Data, ".src.tar.gz"))
		if err != nil {
			return nil, err
		}
		size, err := parseSize(td[4].FirstChild.Data)
		if err != nil {
			return nil, err
		}
		hash, err := hex.DecodeString(tt.FirstChild.Data)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to decode hash string: %s", tt.FirstChild.Data)
		}

		source := Source{
			Version: version,
			HRef:    getAttrValue(a, "href"),
			Size:    size,
			Hash:    hash,
		}
		result = append(result, source)
	}

	sort.Slice(result, func(i, j int) bool {
		return !result[i].Version.Less(result[j].Version)
	})
	return result, nil
}

func findFirstNodeByTag(node *html.Node, tag string) *html.Node {
	result := findNodeByTag(node, tag)
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

func findNodeByTag(node *html.Node, tag string) []*html.Node {
	result := make([]*html.Node, 0)
	if node.Type == html.ElementNode && node.Data == tag {
		result = append(result, node)
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		res := findNodeByTag(child, tag)
		result = append(result, res...)
	}
	return result
}

func findFirstNodeByTagAndClass(node *html.Node, tag string, class string) *html.Node {
	result := findNodeByTagAndClass(node, tag, class)
	if len(result) == 0 {
		return nil
	}
	return result[0]
}

func findNodeByTagAndClass(node *html.Node, tag string, class string) []*html.Node {
	result := make([]*html.Node, 0)
	if node.Type == html.ElementNode && node.Data == tag {
		for _, attr := range node.Attr {
			if attr.Key == "class" {
				if attr.Val == class {
					result = append(result, node)
					break
				}
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		res := findNodeByTagAndClass(child, tag, class)
		result = append(result, res...)
	}

	return result
}

func hasAttr(node *html.Node, key string) bool {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return true
		}
	}
	return false
}

func hasAttrValue(node *html.Node, key string, value string) bool {
	for _, attr := range node.Attr {
		if attr.Key == key && attr.Val == value {
			return true
		}
	}
	return false
}

func getAttrValue(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func parseSize(s string) (int64, error) {
	var ofs int64
	switch {
	case strings.HasSuffix(s, "KB"):
		ofs = 1024
	case strings.HasSuffix(s, "MB"):
		ofs = 1024 * 1024
	case strings.HasSuffix(s, "GB"):
		ofs = 1024 * 1024 * 1024
	default:
		return 0, errors.Newf("invalid format: %s", s)
	}

	// trim
	s = s[:len(s)-2]
	s = strings.TrimSpace(s)

	iv, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "invalid format: %s", s)
	}
	return iv * ofs, nil
}
