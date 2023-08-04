package backup

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
)

type Maker interface {
	Copy() error
	Restore() error
	Close() error
}

func New(dir filepathx.DirPath, origin filepathx.Path) (*Backup, error) {
	if err := dir.Ensure(); err != nil {
		return nil, errorx.Errorf(err, "new backup")
	}
	return &Backup{
		dir:    dir,
		origin: origin,
		path:   dir.Join(origin.Tail()),
	}, nil
}

type Backup struct {
	dir    filepathx.DirPath
	origin filepathx.Path
	path   filepathx.Path
}

func (b *Backup) Restore() error {
	return b.copy(b.origin, b.path)
}

func (b *Backup) Copy() error {
	return b.copy(b.path, b.origin)
}

func (*Backup) copy(dst filepathx.Path, src filepathx.Path) error {
	isDir, err := src.IsDir()
	if err != nil {
		return errorx.Errorf(err, "backup origin")
	}

	if isDir {
		return src.DirPath().Copy(dst.DirPath())
	}
	return src.FilePath().Copy(dst.FilePath())
}

func (b *Backup) Close() error {
	return b.dir.Remove()
}
