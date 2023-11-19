package remote

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
)

// Version はソースコードのバージョンを表します。
type Version struct {
	Major    int    // メジャーバージョン。
	Minor    int    // マイナーバージョン。
	Patch    int    // パッチバージョン。
	RC       int    // リリース候補バージョン。RC以外では0。
	Beta     int    // ベータバージョン。ベータ以外では0。
	Original string // バージョン文字列のオリジナル表現。
}

// Less は Version 同士を比較し、左辺が小さい値を取るかどうかを返します。
func (v Version) Less(rhs Version) bool {
	if v.Major != rhs.Major {
		return v.Major < rhs.Major
	}
	if v.Minor != rhs.Minor {
		return v.Minor < rhs.Minor
	}
	if v.Patch != rhs.Patch {
		return v.Patch < rhs.Patch
	}
	if v.RC != rhs.RC {
		return v.RC < rhs.RC
	}
	if v.Beta != rhs.Beta {
		return v.Beta < rhs.Beta
	}
	return false
}

// String は fmt.Stringer の実装です。
func (v Version) String() string {
	return v.Original
}

// parseVersion はバージョン文字列からバージョン構造体を作成します。
func parseVersion(s string) (Version, error) {
	original := s

	s = strings.TrimPrefix(s, "go")

	elements := strings.Split(s, ".")
	if len(elements) < 2 {
		return Version{}, errors.Newf("invalid format: %s", original)
	}

	parseElem := func(s string) (int, int, error) {
		if !strings.Contains(s, "rc") && !strings.Contains(s, "beta") {
			iv, err := strconv.ParseInt(s, 0, 64)
			if err != nil {
				return 0, 0, errors.Wrapf(err, "invalid format: %s", original)
			}
			return int(iv), 0, nil
		}

		var elems []string
		switch {
		case strings.Contains(s, "rc"):
			elems = strings.Split(s, "rc")
		case strings.Contains(s, "beta"):
			elems = strings.Split(s, "beta")
		default:
			return 0, 0, errors.Newf("invalid format: %s", original)
		}
		if len(elems) != 2 {
			return 0, 0, errors.Newf("invalid format: %s", original)
		}
		iv, err := strconv.ParseInt(elems[0], 0, 64)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "invalid format: %s", original)
		}
		rc, err := strconv.ParseInt(elems[1], 0, 64)
		if err != nil {
			return 0, 0, errors.Wrapf(err, "invalid format: %s", original)
		}

		return int(iv), int(rc), nil
	}

	major, rc0, err := parseElem(elements[0])
	if err != nil {
		return Version{}, err
	}
	minor, rc1, err := parseElem(elements[1])
	if err != nil {
		return Version{}, err
	}

	var patch, rc2 int
	if len(elements) > 2 {
		patch, rc2, err = parseElem(elements[2])
		if err != nil {
			return Version{}, err
		}
	}

	extra := max(rc0, rc1, rc2) // rcやbetaは正しいフォーマットなら1つしか指定されないので、 0, 0, 1 のような値を取ることになり、maxすれば適切な値がピックできる

	version := Version{
		Major:    major,
		Minor:    minor,
		Patch:    patch,
		Original: original,
	}
	switch {
	case strings.Contains(s, "rc"):
		version.RC = extra
	case strings.Contains(s, "beta"):
		version.Beta = extra
	}

	return version, nil
}
