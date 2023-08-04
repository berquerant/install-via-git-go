package filepathx

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Path string

var (
	ErrNotAbs  = errors.New("NotAbs")
	ErrNotDir  = errors.New("NotDir")
	ErrNotFile = errors.New("NotFile")
)

func PWD() DirPath {
	p, _ := NewPath(".")
	return p.DirPath()
}

func NewPath(path string) (Path, error) {
	x, err := filepath.Abs(path)
	if err != nil {
		return Path(""), errors.Join(ErrNotAbs, err)
	}
	return Path(x), nil
}

// Join joins relPath to the path if relPath is not an absolute path.
func (p Path) Join(relPath string) Path {
	if filepath.IsLocal(relPath) {
		return Path(filepath.Join(p.String(), relPath))
	}
	return p
}

// Ext returns the extension without dot.
func (p Path) Ext() (string, bool) {
	ext := filepath.Ext(p.String())
	if ext == "" {
		return "", false
	}
	return strings.TrimLeft(ext, "."), true
}

func (p Path) Parent() Path {
	return Path(filepath.Dir(p.String()))
}

func (p Path) Tail() string {
	xs := strings.Split(p.String(), "/")
	return xs[len(xs)-1]
}

func (p Path) String() string {
	return string(p)
}

func (p Path) DirPath() DirPath {
	return DirPath{p}
}

func (p Path) FilePath() FilePath {
	return FilePath{p}
}

func (p Path) IsDir() (bool, error) {
	info, err := os.Stat(p.String())
	if err != nil {
		return false, err
	}
	return info.IsDir(), nil
}
