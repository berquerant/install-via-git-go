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

func (d DirPath) Copy(dst DirPath) (retErr error) {
	defer func() {
		logx.Debug("copy dir",
			logx.S("src", d.String()),
			logx.S("dst", dst.String()),
			logx.Err(retErr),
		)
	}()

	in, err := os.Open(d.String())
	if err != nil {
		retErr = err
		return
	}
	defer in.Close()

	files, err := in.Readdir(-1)
	if err != nil {
		retErr = err
		return
	}

	if err := os.MkdirAll(dst.String(), 0755); err != nil {
		retErr = err
		return
	}

	for _, file := range files {
		srcPath := d.Join(file.Name())
		dstPath := dst.Join(file.Name())

		if file.IsDir() {
			if err := srcPath.DirPath().Copy(dstPath.DirPath()); err != nil {
				retErr = err
				return
			}
			continue
		}

		if err := srcPath.FilePath().Copy(dstPath.FilePath()); err != nil {
			retErr = err
			return
		}
	}
	return
}

func (d DirPath) Move(dst DirPath) (retErr error) {
	defer func() {
		logx.Debug("move dir",
			logx.S("src", d.String()),
			logx.S("dst", dst.String()),
			logx.Err(retErr),
		)
	}()

	in, err := os.Open(d.String())
	if err != nil {
		retErr = err
		return
	}
	defer in.Close()

	files, err := in.Readdir(-1)
	if err != nil {
		retErr = err
		return
	}

	if err := os.MkdirAll(dst.String(), 0755); err != nil {
		retErr = err
		return
	}

	for _, file := range files {
		srcPath := d.Join(file.Name())
		dstPath := dst.Join(file.Name())

		if file.IsDir() {
			if err := srcPath.DirPath().Move(dstPath.DirPath()); err != nil {
				retErr = err
				return
			}
			continue
		}

		if err := srcPath.FilePath().Move(dstPath.FilePath()); err != nil {
			retErr = err
			return
		}
	}
	return
}
