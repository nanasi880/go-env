package remote

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"io"
	"os"
	"path/filepath"

	"github.com/cockroachdb/errors"
)

// Archive はソースコードのアーカイブです。
type Archive struct {
	fullName string
	source   Source
}

// Extract は指定されたディレクトリにアーカイブを展開します。
func (a Archive) Extract(dir string) error {
	f, err := os.Open(a.fullName)
	if err != nil {
		return errors.Wrap(err, a.fullName)
	}
	defer f.Close()

	h := sha256.New()
	_, err = io.Copy(h, f)
	if err != nil {
		return errors.Wrap(err, a.fullName)
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return errors.Wrap(err, a.fullName)
	}

	sum := h.Sum(nil)
	if !bytes.Equal(a.source.Hash, sum) {
		return errors.New("Checksum mismatch")
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		return errors.Wrap(err, "")
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return errors.Wrap(err, "")
		}
		full := filepath.Join(dir, header.Name)
		perm := header.FileInfo().Mode().Perm()
		switch header.Typeflag {
		case tar.TypeDir:
			err := os.MkdirAll(full, perm)
			if err != nil {
				return errors.Wrap(err, "os.MkdirAll")
			}
		case tar.TypeReg:
			err := func() error {
				out, err := os.OpenFile(full, os.O_CREATE|os.O_RDWR|os.O_TRUNC, perm)
				if err != nil {
					return errors.Wrap(err, "os.OpenFile")
				}
				defer out.Close()
				_, err = io.Copy(out, tr)
				if err != nil {
					return errors.Wrap(err, "")
				}
				return nil
			}()
			if err != nil {
				return err
			}
		default:
			return errors.Newf("not supported tar file type: %d", header.Typeflag)
		}
	}
}
