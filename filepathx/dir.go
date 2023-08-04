package filepathx

import (
	"berquerant/install-via-git-go/logx"
	"os"
)

type DirPath struct {
	Path
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

func (d DirPath) Exist() bool {
	stat, err := os.Stat(d.String())
	return err == nil && stat.IsDir()
}

type WalkCallback func(dst, src Path) error

func (d DirPath) WalkWith(dst DirPath, callback WalkCallback) error {
	in, err := os.Open(d.String())
	if err != nil {
		return err
	}
	defer in.Close()

	files, err := in.Readdir(-1)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst.String(), 0755); err != nil {
		return err
	}

	for _, file := range files {
		srcPath := d.Join(file.Name())
		dstPath := dst.Join(file.Name())
		if err := callback(dstPath, srcPath); err != nil {
			return err
		}
	}
	return nil
}

func (d DirPath) Copy(dst DirPath) (retErr error) {
	return d.WalkWith(dst, func(dst, src Path) error {
		isDir, err := src.IsDir()
		if err != nil {
			return err
		}
		if isDir {
			return src.DirPath().Copy(dst.DirPath())
		}
		return src.FilePath().Copy(dst.FilePath())
	})
}

func (d DirPath) Move(dst DirPath) error {
	return d.WalkWith(dst, func(dst, src Path) error {
		isDir, err := src.IsDir()
		if err != nil {
			return err
		}
		if isDir {
			return src.DirPath().Move(dst.DirPath())
		}
		return src.FilePath().Move(dst.FilePath())
	})
}
