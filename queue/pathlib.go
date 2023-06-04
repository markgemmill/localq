package queue

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Path struct {
	path     string
	fs       afero.Fs
	fileMode os.FileMode
	err      error
}

func NewPath(path string, fs afero.Fs, fileMode os.FileMode) Path {
	return Path{
		path:     path,
		fs:       fs,
		fileMode: fileMode,
	}

}

func (pth Path) Copy() Path {
	p := NewPath(pth.path, pth.fs, pth.fileMode)
	p.err = pth.err
	return p
}

func (pth Path) SetPath(path string) Path {
	p := pth.Copy()
	p.path = path
	return p
}

func (pth Path) SetErr(err error) Path {
	p := pth.Copy()
	p.err = err
	return p
}

func (pth Path) HasError() bool {
	return pth.err != nil
}

func (pth Path) Resolve() Path {
	if pth.HasError() {
		return pth
	}
	absPth, err := filepath.Abs(pth.path)
	if err != nil {
		pth.err = err
		return pth
	}
	return pth.SetErr(err).SetPath(absPth)
}

func (pth Path) Suffix() string {
	return filepath.Ext(pth.path)
}

func (pth Path) Name() string {
	return filepath.Base(pth.path)
}

func (pth Path) Stem() string {
	segments := strings.Split(filepath.Base(pth.path), ".")
	t := len(segments)
	switch t {
	case 1:
		return segments[0]
	case 2:
		return segments[0]
	default:
		return strings.Join(segments[0:t-2], ".")
	}
}

func (pth Path) String() string {
	return pth.path
}

func (pth Path) Exists() bool {
	ok, _ := afero.Exists(pth.fs, pth.path)
	return ok
}

func (pth Path) IsDir() bool {
	ok, _ := afero.IsDir(pth.fs, pth.path)
	return ok
}

func (pth Path) IsFile() bool {
	ok, _ := afero.IsDir(pth.fs, pth.path)
	return !ok
}

func (pth Path) MkDir() error {
	return pth.fs.Mkdir(pth.path, pth.fileMode)
}

func (pth Path) MkDirs() error {
	return pth.fs.MkdirAll(pth.path, pth.fileMode)
}

func (pth Path) Read() ([]byte, error) {
	return afero.ReadFile(pth.fs, pth.path)
}

func (pth Path) Write(data []byte) error {
	return afero.WriteFile(pth.fs, pth.path, data, pth.fileMode)
}

func (pth Path) Join(paths ...string) Path {
	return pth.SetPath(filepath.Join(pth.path, filepath.Join(paths...)))
}

func (pth Path) Parent() Path {
	return pth.SetPath(filepath.Dir(pth.path))
}

func (pth Path) Remove() error {
	return pth.fs.Remove(pth.path)
}

func (pth Path) Stat() (os.FileInfo, error) {
	return pth.fs.Stat(pth.path)
}

func (pth Path) ModTime() (time.Time, error) {
	stat, err := pth.Stat()
	if err != nil {
		return time.Now(), err
	}
	return stat.ModTime(), nil
}

func (pth Path) ReadDir() ([]Path, error) {
	dirToRead := pth.path
	if pth.IsFile() {
		dirToRead = pth.Parent().path
	}
	content, err := afero.ReadDir(pth.fs, dirToRead)
	if err != nil {
		return nil, err
	}
	paths := []Path{}
	directory := pth.SetPath(dirToRead)
	for _, info := range content {
		paths = append(paths, directory.Join(info.Name()))
	}
	return paths, nil
}
