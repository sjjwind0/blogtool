package archive

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
)

func UnZip(zipPath string) error {
	dest := filepath.Dir(zipPath)
	unZipFile, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	os.MkdirAll(dest, 0755)
	for _, f := range unZipFile.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		path := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			_, err = io.Copy(f, rc)
			if err != nil {
				if err != io.EOF {
					return err
				}
			}
			f.Close()
		}
	}
	unZipFile.Close()
	return nil
}
