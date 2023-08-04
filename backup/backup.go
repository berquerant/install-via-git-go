package backup

import (
	"berquerant/install-via-git-go/errorx"
	"berquerant/install-via-git-go/filepathx"
	"os"
)

//go:generate go run github.com/berquerant/goconfig@v0.3.0 -field "Rename bool" -option -output backup_config_generated.go

type Maker interface {
	// Copy copies origin to backup.
	Copy() error
	// Restore moves backup to origin.
	Restore(opt ...ConfigOption) error
	// Close removes backup.
	Close() error
	// Move moves origin to backup.
	Move() error
	// Rename renames origin to backup.
	Rename() error
}

func IntoTempDir(origin filepathx.Path) (*Backup, error) {
	dir, err := os.MkdirTemp("", "install_via_git_backup")
	if err != nil {
		return nil, errorx.Errorf(err, "new backup")
	}
	dirPath, _ := filepathx.NewPath(dir)
	return New(dirPath.DirPath(), origin)
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

func (b *Backup) Restore(opt ...ConfigOption) error {
	config := NewConfigBuilder().Rename(false).Build()
	config.Apply(opt...)
	if config.Rename.Get() {
		return b.path.Rename(b.origin)
	}
	return b.move(b.origin, b.path)
}

func (b *Backup) Rename() error {
	return b.origin.Rename(b.path)
}

func (b *Backup) Move() error {
	return b.move(b.path, b.origin)
}

func (b *Backup) Copy() error {
	return b.copy(b.path, b.origin)
}

func (*Backup) move(dst filepathx.Path, src filepathx.Path) error {
	isDir, err := src.IsDir()
	if err != nil {
		return errorx.Errorf(err, "backup origin")
	}

	if isDir {
		return src.DirPath().Move(dst.DirPath())
	}
	return src.FilePath().Move(dst.FilePath())
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
