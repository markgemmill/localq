package queue

import (
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"testing"
)

func MakePath() Path {
	return NewPath("/some/root", afero.NewMemMapFs(), 0777)
}

func TestNewPath(t *testing.T) {
	pth := MakePath()
	assert.Equal(t, "/some/root", pth.String())
	assert.False(t, pth.HasError())
}

func TestPathCopy(t *testing.T) {
	pth := MakePath()
	pth2 := pth.Copy()
	assert.True(t, pth == pth2)
}

func TestPath_DirInfo(t *testing.T) {
	pth := MakePath()
	assert.Equal(t, "root", pth.Name())
	assert.Equal(t, "", pth.Suffix())
	assert.Equal(t, "root", pth.Stem())

	//assert.Equal(t, false, pth.Exists())
	//assert.Equal(t, false, pth.IsDir())
	//assert.Equal(t, false, pth.IsFile())
}

func TestPath_MkDirs(t *testing.T) {
	pth := MakePath()

	assert.Equal(t, false, pth.Exists())

	err := pth.MkDirs()
	assert.Nil(t, err)

	assert.Equal(t, true, pth.Exists())
}

func TestPath_MkDir(t *testing.T) {
	pth := MakePath()
	pth = pth.Join("foo")

	assert.Equal(t, false, pth.Exists())

	err := pth.MkDir()
	assert.Nil(t, err)

	assert.Equal(t, true, pth.Exists())

}

func TestPath_FileInfo(t *testing.T) {
	pth := MakePath()
	pth = pth.Join("file.txt")
	assert.Equal(t, "file.txt", pth.Name())
	assert.Equal(t, ".txt", pth.Suffix())
	assert.Equal(t, "file", pth.Stem())
}

func TestPath_Write(t *testing.T) {
	pth := MakePath()
	pth = pth.Join("file.txt")

	assert.Equal(t, false, pth.Exists())
	assert.Equal(t, true, pth.IsFile())

	err := pth.Write([]byte("data"))
	assert.Nil(t, err)

	assert.Equal(t, true, pth.Exists())
	assert.Equal(t, true, pth.IsFile())

}
