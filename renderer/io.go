package renderer

import (
	"io"
	"os"
	"path/filepath"
)

func copyFile(src string, dst string) error {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func(r *os.File) {
		_ = r.Close()
	}(r)

	err = os.MkdirAll(filepath.Dir(dst), 0755)
	if err != nil {
		return err
	}

	w, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(w *os.File) {
		_ = w.Close()
	}(w)

	_, err = io.Copy(w, r)
	if err != nil {
		return err
	}

	return nil
}
