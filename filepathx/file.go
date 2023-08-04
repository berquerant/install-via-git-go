package filepathx

import (
	"berquerant/install-via-git-go/logx"
	"io"
	"io/fs"
	"os"
)

type FilePath struct {
	Path
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

func (f FilePath) Copy(dst FilePath) (retErr error) {
	defer func() {
		logx.Debug("copy file",
			logx.S("src", f.String()),
			logx.S("dst", dst.String()),
			logx.Err(retErr),
		)
	}()

	in, err := os.Open(f.String())
	if err != nil {
		retErr = err
		return
	}
	defer in.Close()

	out, err := os.Create(dst.String())
	if err != nil {
		retErr = err
		return
	}
	defer out.Close()

	_, retErr = io.Copy(out, in)
	return
}
