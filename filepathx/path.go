package filepathx

import (
	"berquerant/install-via-git-go/logx"
	"errors"
	"io/fs"
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

type (
	DirPath struct {
		Path
	}
	FilePath struct {
		Path
	}
)

func (p Path) DirPath() DirPath {
	return DirPath{p}
}

func (p Path) FilePath() FilePath {
	return FilePath{p}
}

func (d DirPath) Ensure() error {
	err := os.MkdirAll(d.String(), 0750)
	logx.Debug("ensure dir", logx.S("path", d.String()), logx.Err(err))
	return err
}

func (d DirPath) Remove() error {
	err := os.RemoveAll(d.String())
	logx.Debug("remove dir", logx.S("path", d.String()), logx.Err(err))
	return err
}

func (f FilePath) DirPath() DirPath {
	return f.Parent().DirPath()
}

func (f FilePath) Exist() bool {
	stat, err := os.Stat(f.String())
	return err == nil && !stat.IsDir()
}

func (f FilePath) Ensure() (retErr error) {
	defer func() {
		logx.Debug("ensure file", logx.S("file", f.String()), logx.Err(retErr))
	}()
	if err := f.DirPath().Ensure(); err != nil {
		retErr = err
		return
	}
	var stat fs.FileInfo
	stat, err := os.Stat(f.String())
	if err == nil {
		if stat.IsDir() {
			retErr = ErrNotFile
		}
		return
	}
	if !os.IsNotExist(err) {
		return
	}
	var file *os.File
	if file, err = os.Create(f.String()); err != nil {
		retErr = err
		return
	}
	retErr = file.Close()
	return
}

// Read reads the whole file.
func (f FilePath) Read() (string, error) {
	b, err := os.ReadFile(f.String())
	logx.Debug("read file", logx.S("path", f.String()), logx.S("content", string(b)), logx.Err(err))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Write overwrites the file.
func (f FilePath) Write(str string) error {
	err := os.WriteFile(f.String(), []byte(str), 0600)
	logx.Debug("write file", logx.S("path", f.String()), logx.S("content", str), logx.Err(err))
	return err
}

func (f FilePath) Remove() error {
	err := os.Remove(f.String())
	logx.Debug("remove file", logx.S("path", f.String()), logx.Err(err))
	return err
}
