package main

import (
	"embed"
	"io/fs"

	"github.com/psanford/memfs"
)

func moveThumbnails(imagefs embed.FS, mfs *memfs.FS) error {
	err := fs.WalkDir(imagefs, ".", func(path string, f fs.DirEntry, err error) error {
		if !f.IsDir() {
			buf, err := imagefs.ReadFile(path)
			if err != nil {
				return err
			}
			err = mfs.WriteFile(f.Name(), buf, 0666)
			if err != nil {
				return err
			}
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
