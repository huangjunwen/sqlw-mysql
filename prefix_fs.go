package main

import (
	"net/http"
	"path"
)

// PrefixFS adds a path prefix to a FileSystem.
type PrefixFS struct {
	prefix string
	fs     http.FileSystem
}

func newPrefixFS(prefix string, fs http.FileSystem) *PrefixFS {
	return &PrefixFS{
		prefix: prefix,
		fs:     fs,
	}
}

// Open implements http.FileSystem interface.
func (fs *PrefixFS) Open(name string) (http.File, error) {
	return fs.fs.Open(path.Join(fs.prefix, name))
}
