package utils

import (
	"archive/tar"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func DirectoryExists(directory string) bool {
	if _, err := os.Stat(directory); err == nil {
		return true
	} else {
		return !os.IsNotExist(err)
	}
}

func HTTPErrorCheck(err error, w http.ResponseWriter, errorCode int) bool {
	if err != nil {
		http.Error(w, err.Error(), errorCode)
		return true
	}

	return false
}

func Tar(src string) (string, error) {
	if !DirectoryExists(src) {
		return "", ERRFILENOTFOUND
	}

	archiveName := path.Join(Config.GetArchiveDir(), src) + ".tar"

	dir, err := os.Open(archiveName)

	if err != nil {
		return "", ERRNOOPEN
	}

	tw := tar.NewWriter(dir)
	defer tw.Close()

	return archiveName, filepath.Walk(src, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(strings.Replace(file, src, "", -1), string(filepath.Separator))

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		f, err := os.Open(file)
		if err != nil {
			return err
		}

		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		f.Close()

		return nil
	})
}
