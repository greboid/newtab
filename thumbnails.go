package main

import (
	"bytes"
	"embed"
	"image"
	"image/png"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/psanford/memfs"
	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

func createThumbnails(imagefs embed.FS, mfs *memfs.FS) error {
	err := fs.WalkDir(imagefs, ".", func(path string, f fs.DirEntry, err error) error {
		if !f.IsDir() {
			err = createThumbnail(f.Name(), imagefs, mfs)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func createThumbnail(filename string, imagefs embed.FS, mfs *memfs.FS) error {
	var buf bytes.Buffer
	var m image.Image
	var data []byte
	var err error

	data, err = imagefs.ReadFile("images/" + filename)
	if err != nil {
		return err
	}
	m, _, err = image.Decode(bytes.NewReader(data))
	if err != nil {
		return err
	}
	m = resize.Resize(80, 80, m, resize.Lanczos3)

	err = webp.Encode(&buf, m, &webp.Options{Lossless: false, Quality: 60})
	if err != nil {
		return err
	}
	err = mfs.WriteFile(filepath.Base(filename)+".webp", buf.Bytes(), 0666)
	if err != nil {
		return err
	}
	err = png.Encode(&buf, m)
	if err != nil {
		return err
	}
	err = mfs.WriteFile(strings.TrimSuffix(filename, filepath.Ext(filename))+filepath.Ext(filename), buf.Bytes(), 0666)
	if err != nil {
		return err
	}
	return nil
}
